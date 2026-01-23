package answer

import (
	"encoding/json"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

type oilfieldTemplate struct {
	Level uint32 `json:"level"`
	Time  uint32 `json:"time"`
}

type classUpgradeTemplate struct {
	Level uint32 `json:"level"`
	Time  uint32 `json:"time"`
}

type navalAcademyShoppingTemplate struct {
	SpecialGoodsNum uint32 `json:"special_goods_num"`
}

func ResourcesInfo(buffer *[]byte, client *connection.Client) (int, int, error) {
	response := protobuf.SC_22001{
		OilWellLevel:       proto.Uint32(1),
		OilWellLvUpTime:    proto.Uint32(0),
		GoldWellLevel:      proto.Uint32(1),
		GoldWellLvUpTime:   proto.Uint32(0),
		ClassLv:            proto.Uint32(1),
		ClassLvUpTime:      proto.Uint32(0),
		SkillClassNum:      proto.Uint32(0),
		DailyFinishBuffCnt: proto.Uint32(0),
		Class: &protobuf.NAVALACADEMY_CLASS{
			Proficiency: proto.Uint32(0),
		},
	}
	oilfieldEntries, err := orm.ListConfigEntries(orm.GormDB, "ShareCfg/oilfield_template.json")
	if err != nil {
		return 0, 22001, err
	}
	if len(oilfieldEntries) > 0 {
		var template oilfieldTemplate
		if err := json.Unmarshal(oilfieldEntries[0].Data, &template); err != nil {
			return 0, 22001, err
		}
		response.OilWellLevel = proto.Uint32(template.Level)
		response.OilWellLvUpTime = proto.Uint32(template.Time)
		response.GoldWellLevel = proto.Uint32(template.Level)
		response.GoldWellLvUpTime = proto.Uint32(template.Time)
	}
	classEntries, err := orm.ListConfigEntries(orm.GormDB, "ShareCfg/class_upgrade_template.json")
	if err != nil {
		return 0, 22001, err
	}
	if len(classEntries) > 0 {
		var template classUpgradeTemplate
		if err := json.Unmarshal(classEntries[0].Data, &template); err != nil {
			return 0, 22001, err
		}
		response.ClassLv = proto.Uint32(template.Level)
		response.ClassLvUpTime = proto.Uint32(template.Time)
	}
	academyEntries, err := orm.ListConfigEntries(orm.GormDB, "ShareCfg/navalacademy_data_template.json")
	if err != nil {
		return 0, 22001, err
	}
	if len(academyEntries) > 0 {
		response.SkillClassNum = proto.Uint32(uint32(len(academyEntries)))
	}
	shoppingEntries, err := orm.ListConfigEntries(orm.GormDB, "ShareCfg/navalacademy_shoppingstreet_template.json")
	if err != nil {
		return 0, 22001, err
	}
	if len(shoppingEntries) > 0 {
		var template navalAcademyShoppingTemplate
		if err := json.Unmarshal(shoppingEntries[0].Data, &template); err != nil {
			return 0, 22001, err
		}
		response.DailyFinishBuffCnt = proto.Uint32(template.SpecialGoodsNum)
	}
	return client.SendMessage(22001, &response)
}
