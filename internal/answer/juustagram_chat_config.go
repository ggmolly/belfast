package answer

import (
	"encoding/json"
	"fmt"

	"github.com/ggmolly/belfast/internal/orm"
)

const (
	juustagramChatGroupConfigCategory = "ShareCfg/activity_ins_chat_group.json"
	juustagramRedPacketConfigCategory = "ShareCfg/activity_ins_redpackage.json"
)

type juustagramChatGroupConfig struct {
	ID        uint32 `json:"id"`
	ShipGroup uint32 `json:"ship_group"`
}

type juustagramRedPacketConfig struct {
	ID      uint32   `json:"id"`
	Type    uint32   `json:"type"`
	Content []uint32 `json:"content"`
}

func getJuustagramChatGroupConfig(chatGroupID uint32) (*juustagramChatGroupConfig, error) {
	entry, err := orm.GetConfigEntry(orm.GormDB, juustagramChatGroupConfigCategory, fmt.Sprintf("%d", chatGroupID))
	if err != nil {
		return nil, err
	}
	var config juustagramChatGroupConfig
	if err := json.Unmarshal(entry.Data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func getJuustagramRedPacketConfig(redPacketID uint32) (*juustagramRedPacketConfig, error) {
	entry, err := orm.GetConfigEntry(orm.GormDB, juustagramRedPacketConfigCategory, fmt.Sprintf("%d", redPacketID))
	if err != nil {
		return nil, err
	}
	var config juustagramRedPacketConfig
	if err := json.Unmarshal(entry.Data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
