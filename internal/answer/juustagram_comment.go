package answer

import (
	"errors"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/consts"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

func JuustagramComment(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_11703
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, consts.JuustagramPacketCommentResp, err
	}
	if client.Commander == nil {
		return 0, consts.JuustagramPacketCommentResp, errors.New("missing commander")
	}
	now := uint32(time.Now().Unix())
	option, err := ensureJuustagramOption(payload.GetId(), payload.GetDiscuss(), payload.GetIndex())
	if err != nil {
		return 0, consts.JuustagramPacketCommentResp, err
	}
	entry, err := orm.GetJuustagramPlayerDiscuss(client.Commander.CommanderID, payload.GetId(), payload.GetDiscuss())
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return 0, consts.JuustagramPacketCommentResp, err
		}
		entry = &orm.JuustagramPlayerDiscuss{}
	}
	entry.CommanderID = client.Commander.CommanderID
	entry.MessageID = payload.GetId()
	entry.DiscussID = payload.GetDiscuss()
	entry.OptionIndex = payload.GetIndex()
	entry.NpcReplyID = option.NpcReplyID
	entry.CommentTime = now
	if err := orm.UpsertJuustagramPlayerDiscuss(entry); err != nil {
		return 0, consts.JuustagramPacketCommentResp, err
	}
	message, err := BuildJuustagramMessage(client.Commander.CommanderID, payload.GetId(), now)
	if err != nil {
		return 0, consts.JuustagramPacketCommentResp, err
	}
	response := protobuf.SC_11704{
		Result: proto.Uint32(0),
		Data:   message,
	}
	return client.SendMessage(consts.JuustagramPacketCommentResp, &response)
}
