package answer

import (
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/consts"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func AttireApply(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_11005
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 11006, err
	}
	response := protobuf.SC_11006{Result: proto.Uint32(0)}
	attireType := payload.GetType()
	attireID := payload.GetId()
	if attireType != consts.AttireTypeIconFrame && attireType != consts.AttireTypeChatFrame && attireType != consts.AttireTypeCombatUI {
		response.Result = proto.Uint32(1)
		return client.SendMessage(11006, &response)
	}
	if attireID != 0 {
		owned, err := orm.CommanderHasAttire(client.Commander.CommanderID, attireType, attireID, time.Now())
		if err != nil {
			return 0, 11006, err
		}
		if !owned {
			response.Result = proto.Uint32(2)
			return client.SendMessage(11006, &response)
		}
	}
	switch attireType {
	case consts.AttireTypeIconFrame:
		client.Commander.SelectedIconFrameID = attireID
		if err := orm.GormDB.Model(client.Commander).Update("selected_icon_frame_id", attireID).Error; err != nil {
			response.Result = proto.Uint32(1)
		}
	case consts.AttireTypeChatFrame:
		client.Commander.SelectedChatFrameID = attireID
		if err := orm.GormDB.Model(client.Commander).Update("selected_chat_frame_id", attireID).Error; err != nil {
			response.Result = proto.Uint32(1)
		}
	case consts.AttireTypeCombatUI:
		client.Commander.SelectedBattleUIID = attireID
		if err := orm.GormDB.Model(client.Commander).Update("selected_battle_ui_id", attireID).Error; err != nil {
			response.Result = proto.Uint32(1)
		}
	}
	return client.SendMessage(11006, &response)
}
