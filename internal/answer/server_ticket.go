package answer

import (
	"fmt"
	"strconv"
	"strings"
)

const serverTicketPrefix = "=*=*=*=BELFAST=*=*=*="

func formatServerTicket(arg2 uint32) string {
	if arg2 == 0 {
		return serverTicketPrefix
	}
	// embed arg2 so later connections can recover account identity
	return fmt.Sprintf("%s:%d", serverTicketPrefix, arg2)
}

func parseServerTicket(ticket string) uint32 {
	if !strings.HasPrefix(ticket, serverTicketPrefix+":") {
		return 0
	}
	// parse arg2 from the suffix, if present
	value := strings.TrimPrefix(ticket, serverTicketPrefix+":")
	arg2, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return 0
	}
	return uint32(arg2)
}
