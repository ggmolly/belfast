package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func CommanderGuildChat(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_60100
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 60101, err
	}
	const chatLogMaxCount = 100
	count := int(payload.GetCount())
	if count <= 0 {
		count = chatLogMaxCount
	}
	if count > chatLogMaxCount {
		count = chatLogMaxCount
	}
	entries, err := orm.ListGuildChatMessages(guildChatPlaceholderID, count)
	if err != nil {
		return 0, 60101, err
	}
	chatList := make([]*protobuf.GUIDE_CHAT, 0, len(entries))
	for _, entry := range entries {
		chatList = append(chatList, &protobuf.GUIDE_CHAT{
			Player:  buildGuildChatPlayer(&entry.Sender),
			Content: proto.String(entry.Content),
			Time:    proto.Uint32(uint32(entry.SentAt.Unix())),
		})
	}
	response := protobuf.SC_60101{ChatList: chatList}
	return client.SendMessage(60101, &response)
}
