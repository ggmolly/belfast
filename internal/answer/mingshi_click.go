package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func ClickMingShi(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_11506
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 11507, err
	}
	client.Commander.AccPayLv += 5
	if err := orm.GormDB.Model(&orm.Commander{}).
		Where("commander_id = ?", client.Commander.CommanderID).
		Update("acc_pay_lv", client.Commander.AccPayLv).Error; err != nil {
		return 0, 11507, err
	}
	response := protobuf.SC_11507{
		Result: proto.Uint32(0),
	}
	return client.SendMessage(11507, &response)
}
