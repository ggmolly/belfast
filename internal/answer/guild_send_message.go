package answer

import (
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func GuildSendMessage(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_60007
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 60008, err
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
