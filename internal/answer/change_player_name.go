package answer

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/ggmolly/belfast/internal/config"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/consts"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

const (
	changePlayerNameMin = 4
	changePlayerNameMax = 14
)

type gamesetEntry struct {
	KeyValue    uint32          `json:"key_value"`
	Description json.RawMessage `json:"description"`
}

func loadGameSetEntry(key string) (*gamesetEntry, error) {
	entry, err := orm.GetConfigEntry(orm.GormDB, "ShareCfg/gameset.json", key)
	if err != nil {
		return nil, err
	}
	var payload gamesetEntry
	if err := json.Unmarshal(entry.Data, &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}

func ChangePlayerName(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_11007
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 11008, err
	}
	response := protobuf.SC_11008{Result: proto.Uint32(0)}
	name := strings.TrimSpace(payload.GetName())
	if name == "" {
		response.Result = proto.Uint32(1)
		return client.SendMessage(11008, &response)
	}
	if name == client.Commander.Name {
		response.Result = proto.Uint32(1)
		return client.SendMessage(11008, &response)
	}
	nameLength := utf8.RuneCountInString(name)
	if nameLength < changePlayerNameMin || nameLength > changePlayerNameMax {
		response.Result = proto.Uint32(1)
		return client.SendMessage(11008, &response)
	}
	createConfig := config.Current().CreatePlayer
	if len(createConfig.NameBlacklist) > 0 {
		lowerName := strings.ToLower(name)
		for _, blocked := range createConfig.NameBlacklist {
			blocked = strings.TrimSpace(blocked)
			if blocked == "" {
				continue
			}
			if strings.Contains(lowerName, strings.ToLower(blocked)) {
				response.Result = proto.Uint32(1)
				return client.SendMessage(11008, &response)
			}
		}
	}
	if createConfig.NameIllegalPattern != "" {
		matcher, err := regexp.Compile(createConfig.NameIllegalPattern)
		if err != nil {
			return 0, 11008, err
		}
		if matcher.MatchString(name) {
			response.Result = proto.Uint32(1)
			return client.SendMessage(11008, &response)
		}
	}
	var existing orm.Commander
	if err := orm.GormDB.Where("name = ? AND commander_id <> ?", name, client.Commander.CommanderID).First(&existing).Error; err == nil {
		response.Result = proto.Uint32(2015)
		return client.SendMessage(11008, &response)
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, 11008, err
	}

	changeType := payload.GetType()
	if changeType == 0 {
		changeType = 1
	}
	if changeType == 1 {
		levelEntry, err := loadGameSetEntry("player_name_change_lv_limit")
		if err != nil {
			response.Result = proto.Uint32(1)
			return client.SendMessage(11008, &response)
		}
		if client.Commander.Level < int(levelEntry.KeyValue) {
			response.Result = proto.Uint32(1)
			return client.SendMessage(11008, &response)
		}
		cooldownEntry, err := loadGameSetEntry("player_name_cold_time")
		if err != nil {
			response.Result = proto.Uint32(1)
			return client.SendMessage(11008, &response)
		}
		now := time.Now()
		if client.Commander.NameChangeCooldown.After(now) {
			response.Result = proto.Uint32(1)
			return client.SendMessage(11008, &response)
		}
		costEntry, err := loadGameSetEntry("player_name_change_cost")
		if err != nil {
			response.Result = proto.Uint32(1)
			return client.SendMessage(11008, &response)
		}
		var cost []uint32
		if err := json.Unmarshal(costEntry.Description, &cost); err != nil || len(cost) < 3 {
			response.Result = proto.Uint32(1)
			return client.SendMessage(11008, &response)
		}
		costType := cost[0]
		costID := cost[1]
		costCount := cost[2]
		switch costType {
		case 1:
			if !client.Commander.HasEnoughResource(costID, costCount) {
				response.Result = proto.Uint32(1)
				return client.SendMessage(11008, &response)
			}
			if err := client.Commander.ConsumeResource(costID, costCount); err != nil {
				response.Result = proto.Uint32(1)
				return client.SendMessage(11008, &response)
			}
		case 2:
			if !client.Commander.HasEnoughItem(costID, costCount) {
				response.Result = proto.Uint32(1)
				return client.SendMessage(11008, &response)
			}
			if err := client.Commander.ConsumeItem(costID, costCount); err != nil {
				response.Result = proto.Uint32(1)
				return client.SendMessage(11008, &response)
			}
		default:
			response.Result = proto.Uint32(1)
			return client.SendMessage(11008, &response)
		}
		client.Commander.NameChangeCooldown = now.Add(time.Duration(cooldownEntry.KeyValue) * time.Second)
		updates := map[string]interface{}{
			"name":                 name,
			"name_change_cooldown": client.Commander.NameChangeCooldown,
		}
		if err := orm.GormDB.Model(client.Commander).Updates(updates).Error; err != nil {
			response.Result = proto.Uint32(1)
			return client.SendMessage(11008, &response)
		}
		client.Commander.Name = name
		return client.SendMessage(11008, &response)
	}
	if changeType == 2 {
		client.Commander.Name = name
		if err := orm.GormDB.Model(client.Commander).Update("name", name).Error; err != nil {
			response.Result = proto.Uint32(1)
			return client.SendMessage(11008, &response)
		}
		if err := orm.ClearCommanderCommonFlag(orm.GormDB, client.Commander.CommanderID, consts.IllegalityPlayerName); err != nil {
			response.Result = proto.Uint32(1)
		}
		return client.SendMessage(11008, &response)
	}
	response.Result = proto.Uint32(1)
	return client.SendMessage(11008, &response)
}
