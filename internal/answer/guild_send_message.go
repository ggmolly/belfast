package answer

import (
	"fmt"
	"strings"
	"time"

	"github.com/ggmolly/belfast/internal/auth"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

const (
	registrationChatPrefix     = "B-"
	registrationChatRateLimit  = 5
	registrationChatRateWindow = time.Minute
)

var registrationChatLimiter = auth.NewRateLimiter()

func GuildSendMessage(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_60007
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 60008, err
	}
	if pin, ok := extractRegistrationPin(payload.GetChat()); ok {
		commanderID := client.Commander.CommanderID
		key := fmt.Sprintf("registration:%d", commanderID)
		if registrationChatLimiter.Allow(key, registrationChatRateLimit, registrationChatRateWindow) {
			if account, err := orm.ConsumeUserRegistrationChallenge(commanderID, pin, time.Now().UTC()); err == nil {
				auth.LogUserAudit("registration.consume", &account.ID, &commanderID, nil)
			}
		}
		return 0, 60008, nil
	}
	now := time.Now().UTC()
	entry, err := orm.CreateGuildChatMessage(guildChatPlaceholderID, client.Commander.CommanderID, payload.GetChat(), now)
	if err != nil {
		return 0, 60008, err
	}
	chat := &protobuf.GUIDE_CHAT{
		Player:  buildGuildChatPlayer(client.Commander),
		Content: proto.String(entry.Content),
		Time:    proto.Uint32(uint32(entry.SentAt.Unix())),
	}
	packet := &protobuf.SC_60008{Chat: chat}
	client.Server.BroadcastGuildChat(packet)
	return 0, 60008, nil
}

func extractRegistrationPin(message string) (string, bool) {
	trimmed := strings.TrimSpace(message)
	if !strings.HasPrefix(trimmed, registrationChatPrefix) {
		return "", false
	}
	pin := strings.TrimPrefix(trimmed, registrationChatPrefix)
	if len(pin) != 6 {
		return "", false
	}
	for _, ch := range pin {
		if ch < '0' || ch > '9' {
			return "", false
		}
	}
	return pin, true
}
