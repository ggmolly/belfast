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

const collectionTemplateCategory = "ShareCfg/collection_template.json"

type collectionTemplate struct {
	ID          uint32          `json:"id"`
	Exp         uint32          `json:"exp"`
	CollectTime uint32          `json:"collect_time"`
	ShipNum     uint32          `json:"ship_num"`
	ShipLv      uint32          `json:"ship_lv"`
	ShipType    []uint32        `json:"ship_type"`
	Oil         uint32          `json:"oil"`
	DropOilMax  uint32          `json:"drop_oil_max"`
	DropGoldMax uint32          `json:"drop_gold_max"`
	OverTime    uint32          `json:"over_time"`
	DropDisplay json.RawMessage `json:"drop_display"`
	SpecialDrop json.RawMessage `json:"special_drop"`
	Type        uint32          `json:"type"`
	MaxTeam     uint32          `json:"max_team"`
}

func loadCollectionTemplate(collectionID uint32) (*collectionTemplate, error) {
	entry, err := orm.GetConfigEntry(orm.GormDB, collectionTemplateCategory, fmt.Sprintf("%d", collectionID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	var template collectionTemplate
	if err := json.Unmarshal(entry.Data, &template); err != nil {
		return nil, err
	}
	return &template, nil
}

func EventCollectionStart(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_13003
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 13004, err
	}

	response := protobuf.SC_13004{Result: proto.Uint32(1)}
	collectionID := payload.GetId()
	shipIDs := payload.GetShipIdList()
	if collectionID == 0 {
		return client.SendMessage(13004, &response)
	}

	template, err := loadCollectionTemplate(collectionID)
	if err != nil {
		return 0, 13004, err
	}
	if template == nil {
		return client.SendMessage(13004, &response)
	}

	if len(shipIDs) == 0 {
		return client.SendMessage(13004, &response)
	}
	if template.ShipNum > 0 && uint32(len(shipIDs)) != template.ShipNum {
		return client.SendMessage(13004, &response)
	}

	allowedTypes := make(map[uint32]struct{}, len(template.ShipType))
	for _, value := range template.ShipType {
		allowedTypes[value] = struct{}{}
	}
	var hasLevelRequirement bool
	var meetsLevelRequirement bool
	if template.ShipLv > 0 {
		hasLevelRequirement = true
	}
	for _, shipID := range shipIDs {
		owned, ok := client.Commander.OwnedShipsMap[shipID]
		if !ok {
			return client.SendMessage(13004, &response)
		}
		if hasLevelRequirement && owned.Level >= template.ShipLv {
			meetsLevelRequirement = true
		}
		if len(allowedTypes) > 0 {
			if _, ok := allowedTypes[owned.Ship.Type]; !ok {
				return client.SendMessage(13004, &response)
			}
		}
	}
	if hasLevelRequirement && !meetsLevelRequirement {
		return client.SendMessage(13004, &response)
	}

	if template.Oil > 0 && !client.Commander.HasEnoughResource(2, template.Oil) {
		return client.SendMessage(13004, &response)
	}
	if template.DropGoldMax > 0 && client.Commander.GetResourceCount(1) >= template.DropGoldMax {
		return client.SendMessage(13004, &response)
	}
	if template.DropOilMax > 0 && client.Commander.GetResourceCount(2) >= template.DropOilMax {
		return client.SendMessage(13004, &response)
	}
	serverTime := uint32(time.Now().Unix())
	if template.OverTime > 0 && serverTime >= template.OverTime {
		return client.SendMessage(13004, &response)
	}

	finishTime := serverTime + template.CollectTime
	if err := orm.GormDB.Transaction(func(tx *gorm.DB) error {
		if template.MaxTeam > 0 {
			count, err := orm.GetActiveEventCount(tx, client.Commander.CommanderID)
			if err != nil {
				return err
			}
			if uint32(count) >= template.MaxTeam {
				return nil
			}
		}
		existing, err := orm.GetEventCollection(tx, client.Commander.CommanderID, collectionID)
		if err == nil {
			if existing.FinishTime != 0 {
				return nil
			}
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		busy, err := orm.GetBusyEventShipIDs(tx, client.Commander.CommanderID)
		if err != nil {
			return err
		}
		for _, shipID := range shipIDs {
			if _, ok := busy[shipID]; ok {
				return nil
			}
		}

		if template.Oil > 0 {
			if err := client.Commander.ConsumeResourceTx(tx, 2, template.Oil); err != nil {
				return nil
			}
		}
		if existing != nil {
			existing.StartTime = serverTime
			existing.FinishTime = finishTime
			existing.ShipIDs = orm.ToInt64List(shipIDs)
			if err := tx.Save(existing).Error; err != nil {
				return err
			}
		} else {
			event := orm.EventCollection{
				CommanderID:  client.Commander.CommanderID,
				CollectionID: collectionID,
				StartTime:    serverTime,
				FinishTime:   finishTime,
				ShipIDs:      orm.ToInt64List(shipIDs),
			}
			if err := tx.Create(&event).Error; err != nil {
				return err
			}
		}
		response.Result = proto.Uint32(0)
		return nil
	}); err != nil {
		return 0, 13004, err
	}

	if response.GetResult() == 0 {
		update := protobuf.SC_13011{Collection: []*protobuf.COLLECTIONINFO{&protobuf.COLLECTIONINFO{
			Id:         proto.Uint32(collectionID),
			FinishTime: proto.Uint32(finishTime),
			OverTime:   proto.Uint32(template.OverTime),
			ShipIdList: shipIDs,
		}}}
		if _, _, err := client.SendMessage(13011, &update); err != nil {
			return 0, 13004, err
		}
	}
	return client.SendMessage(13004, &response)
}
