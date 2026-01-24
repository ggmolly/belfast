package answer

import (
	"errors"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/consts"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func JuustagramMessageRange(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_11705
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, consts.JuustagramPacketRangeResp, err
	}
	if client.Commander == nil {
		return 0, consts.JuustagramPacketRangeResp, errors.New("missing commander")
	}
	indexBegin := payload.GetIndexBegin()
	indexEnd := payload.GetIndexEnd()
	var templates []orm.JuustagramTemplate
	if err := orm.GormDB.Order("id asc").Where("id >= ? AND id <= ?", indexBegin, indexEnd).Find(&templates).Error; err != nil {
		return 0, consts.JuustagramPacketRangeResp, err
	}
	now := uint32(time.Now().Unix())
	messages := make([]*protobuf.INS_MESSAGE, 0, len(templates))
	for _, template := range templates {
		if ok, err := isPublishableJuustagramTemplate(template); err != nil {
			return 0, consts.JuustagramPacketRangeResp, err
		} else if !ok {
			// Skip templates that are not ready to be served to clients.
			logJuustagramSkip(template, client.Commander.CommanderID)
			continue
		}
		message, err := BuildJuustagramMessage(client.Commander.CommanderID, template.ID, now)
		if err != nil {
			return 0, consts.JuustagramPacketRangeResp, err
		}
		messages = append(messages, message)
	}
	response := protobuf.SC_11706{InsMessageList: messages}
	return client.SendMessage(consts.JuustagramPacketRangeResp, &response)
}
