package answer

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

const serverTicketPrefix = "=*=*=*=BELFAST=*=*=*="

const serverTicketTTL = 15 * time.Minute

type issuedServerTicket struct {
	arg2      uint32
	expiresAt time.Time
}

var serverTickets = struct {
	mu      sync.RWMutex
	entries map[string]issuedServerTicket
}{
	entries: map[string]issuedServerTicket{},
}

func formatServerTicket(arg2 uint32) string {
	if arg2 == 0 {
		return serverTicketPrefix
	}
	token, err := newServerTicketToken()
	if err != nil {
		// keep auth working if random source is unavailable
		return fmt.Sprintf("%s:%d", serverTicketPrefix, arg2)
	}

	now := time.Now()
	serverTickets.mu.Lock()
	pruneExpiredServerTicketsLocked(now)
	serverTickets.entries[token] = issuedServerTicket{arg2: arg2, expiresAt: now.Add(serverTicketTTL)}
	serverTickets.mu.Unlock()

	return fmt.Sprintf("%s:%s", serverTicketPrefix, token)
}

func parseServerTicket(ticket string) uint32 {
	if ticket == serverTicketPrefix {
		return 0
	}
	if !strings.HasPrefix(ticket, serverTicketPrefix+":") {
		return 0
	}

	value := strings.TrimPrefix(ticket, serverTicketPrefix+":")
	if value == "" {
		return 0
	}
	if arg2 := consumeServerTicket(value); arg2 != 0 {
		return arg2
	}

	// compatibility path for older numeric tickets.
	arg2, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return 0
	}
	return uint32(arg2)
}

func newServerTicketToken() (string, error) {
	raw := make([]byte, 24)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func consumeServerTicket(token string) uint32 {
	now := time.Now()

	serverTickets.mu.RLock()
	entry, ok := serverTickets.entries[token]
	serverTickets.mu.RUnlock()
	if !ok {
		return 0
	}
	if now.After(entry.expiresAt) {
		serverTickets.mu.Lock()
		if latest, exists := serverTickets.entries[token]; exists && now.After(latest.expiresAt) {
			delete(serverTickets.entries, token)
		}
		serverTickets.mu.Unlock()
		return 0
	}
	return entry.arg2
}

func pruneExpiredServerTicketsLocked(now time.Time) {
	for token, entry := range serverTickets.entries {
		if now.After(entry.expiresAt) {
			delete(serverTickets.entries, token)
		}
	}
}
