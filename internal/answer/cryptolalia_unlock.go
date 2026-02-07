package answer

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

type soundStoryTemplateEntry struct {
	ID    uint32          `json:"id"`
	Cost1 []uint32        `json:"cost1"`
	Cost2 []uint32        `json:"cost2"`
	Time  json.RawMessage `json:"time"`
}

func CryptolaliaUnlock(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_16205
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 16206, err
	}

	response := protobuf.SC_16206{Ret: proto.Uint32(0)}
	entry, ok, err := loadSoundStoryTemplateEntry(payload.GetId())
	if err != nil {
		return 0, 16206, err
	}
	if !ok {
		response.Ret = proto.Uint32(1)
		return client.SendMessage(16206, &response)
	}
	if !soundStoryTimeAllowsUnlock(entry.Time, time.Now()) {
		response.Ret = proto.Uint32(2)
		return client.SendMessage(16206, &response)
	}

	cost, ok := selectSoundStoryCost(entry, payload.GetCostType())
	if !ok {
		response.Ret = proto.Uint32(1)
		return client.SendMessage(16206, &response)
	}

	const (
		retInvalid      = 1
		retOutOfWindow  = 2
		retInsufficient = 3
		retDBError      = 4
	)

	errInsufficient := errors.New("insufficient")

	commanderID := client.Commander.CommanderID
	err = orm.GormDB.Transaction(func(tx *gorm.DB) error {
		unlocked, err := orm.IsCommanderSoundStoryUnlockedTx(tx, commanderID, payload.GetId())
		if err != nil {
			return err
		}
		if unlocked {
			return nil
		}

		switch cost.dropType {
		case 1:
			if !client.Commander.HasEnoughResource(cost.id, cost.amount) {
				return errInsufficient
			}
			if err := client.Commander.ConsumeResourceTx(tx, cost.id, cost.amount); err != nil {
				return err
			}
		case 2:
			if !client.Commander.HasEnoughItem(cost.id, cost.amount) {
				return errInsufficient
			}
			if err := client.Commander.ConsumeItemTx(tx, cost.id, cost.amount); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported sound story cost type %d", cost.dropType)
		}
		return orm.UnlockCommanderSoundStoryTx(tx, commanderID, payload.GetId())
	})
	if err != nil {
		switch {
		case errors.Is(err, errInsufficient):
			response.Ret = proto.Uint32(retInsufficient)
		case errors.Is(err, gorm.ErrRecordNotFound):
			response.Ret = proto.Uint32(retInvalid)
		default:
			response.Ret = proto.Uint32(retDBError)
		}
		return client.SendMessage(16206, &response)
	}

	return client.SendMessage(16206, &response)
}

type soundStoryCost struct {
	dropType uint32
	id       uint32
	amount   uint32
}

func selectSoundStoryCost(entry *soundStoryTemplateEntry, costType uint32) (soundStoryCost, bool) {
	var cost []uint32
	switch costType {
	case 1:
		cost = entry.Cost1
	case 2:
		cost = entry.Cost2
	default:
		return soundStoryCost{}, false
	}
	if len(cost) != 3 {
		return soundStoryCost{}, false
	}
	return soundStoryCost{dropType: cost[0], id: cost[1], amount: cost[2]}, true
}

func loadSoundStoryTemplateEntry(id uint32) (*soundStoryTemplateEntry, bool, error) {
	key := fmt.Sprintf("%d", id)
	if entry, err := orm.GetConfigEntry(orm.GormDB, "ShareCfg/soundstory_template.json", key); err == nil {
		var out soundStoryTemplateEntry
		if err := json.Unmarshal(entry.Data, &out); err != nil {
			return nil, false, err
		}
		return &out, true, nil
	}
	entries, err := orm.ListConfigEntries(orm.GormDB, "ShareCfg/soundstory_template.json")
	if err != nil {
		return nil, false, err
	}
	for i := range entries {
		var out soundStoryTemplateEntry
		if err := json.Unmarshal(entries[i].Data, &out); err != nil {
			return nil, false, err
		}
		if out.ID == id {
			return &out, true, nil
		}
	}
	return nil, false, nil
}

func soundStoryTimeAllowsUnlock(raw json.RawMessage, now time.Time) bool {
	if len(raw) == 0 {
		return false
	}
	var label string
	if err := json.Unmarshal(raw, &label); err == nil {
		return label == "always"
	}
	var timer any
	if err := json.Unmarshal(raw, &timer); err != nil {
		return false
	}
	data, ok := timer.([]any)
	if !ok || len(data) < 3 {
		return false
	}
	name, ok := data[0].(string)
	if !ok || name != "timer" {
		return false
	}
	start, ok := parseSoundStoryTimerTimestamp(data[1])
	if !ok {
		return false
	}
	end, ok := parseSoundStoryTimerTimestamp(data[2])
	if !ok {
		return false
	}
	return !now.Before(start) && !now.After(end)
}

func parseSoundStoryTimerTimestamp(raw any) (time.Time, bool) {
	parts, ok := raw.([]any)
	if !ok || len(parts) != 2 {
		return time.Time{}, false
	}
	date, ok := parts[0].([]any)
	if !ok || len(date) != 3 {
		return time.Time{}, false
	}
	clock, ok := parts[1].([]any)
	if !ok || len(clock) != 3 {
		return time.Time{}, false
	}
	year, ok := parseJSONInt(date[0])
	if !ok {
		return time.Time{}, false
	}
	month, ok := parseJSONInt(date[1])
	if !ok {
		return time.Time{}, false
	}
	day, ok := parseJSONInt(date[2])
	if !ok {
		return time.Time{}, false
	}
	hour, ok := parseJSONInt(clock[0])
	if !ok {
		return time.Time{}, false
	}
	minute, ok := parseJSONInt(clock[1])
	if !ok {
		return time.Time{}, false
	}
	second, ok := parseJSONInt(clock[2])
	if !ok {
		return time.Time{}, false
	}
	return time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC), true
}
