package answer

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/consts"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

const storeupDataTemplateCategory = "ShareCfg/storeup_data_template.json"

type storeupDataTemplate struct {
	ID           uint32     `json:"id"`
	CharList     []uint32   `json:"char_list"`
	Level        []uint32   `json:"level"`
	AwardDisplay [][]uint32 `json:"award_display"`
}

type storeupGroupStar struct {
	GroupID uint32 `gorm:"column:group_id"`
	MaxStar uint32 `gorm:"column:max_star"`
}

func CollectionGetAward17005(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_17005
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 17006, err
	}

	response := protobuf.SC_17006{Result: proto.Uint32(0)}

	const (
		resultInvalid     = 1
		resultNotEligible = 2
		resultWrongTier   = 3
		resultUnsupported = 4
		resultDBError     = 5
	)

	if err := ensureCommanderLoaded(client, "Collection/Award"); err != nil {
		response.Result = proto.Uint32(resultDBError)
		return client.SendMessage(17006, &response)
	}

	storeupID := payload.GetId()
	awardIndex := payload.GetAwardIndex()
	if storeupID == 0 || awardIndex == 0 {
		response.Result = proto.Uint32(resultInvalid)
		return client.SendMessage(17006, &response)
	}

	template, ok, err := loadStoreupDataTemplate(storeupID)
	if err != nil {
		response.Result = proto.Uint32(resultDBError)
		return client.SendMessage(17006, &response)
	}
	if !ok {
		response.Result = proto.Uint32(resultInvalid)
		return client.SendMessage(17006, &response)
	}
	if int(awardIndex) > len(template.AwardDisplay) || int(awardIndex) > len(template.Level) {
		response.Result = proto.Uint32(resultInvalid)
		return client.SendMessage(17006, &response)
	}

	starCount, err := storeupStarCount(client.Commander.CommanderID, template.CharList)
	if err != nil {
		response.Result = proto.Uint32(resultDBError)
		return client.SendMessage(17006, &response)
	}

	drop := template.AwardDisplay[awardIndex-1]
	if len(drop) < 3 {
		response.Result = proto.Uint32(resultInvalid)
		return client.SendMessage(17006, &response)
	}
	dropType, dropID, dropCount := drop[0], drop[1], drop[2]
	if dropCount == 0 {
		response.Result = proto.Uint32(resultInvalid)
		return client.SendMessage(17006, &response)
	}

	sentinelWrongTier := errors.New("wrong tier")
	sentinelNotEligible := errors.New("not eligible")
	sentinelUnsupported := errors.New("unsupported")

	commanderID := client.Commander.CommanderID
	now := uint32(time.Now().Unix())
	err = orm.GormDB.Transaction(func(tx *gorm.DB) error {
		if template.Level[awardIndex-1] > starCount {
			return sentinelNotEligible
		}
		advanced, err := orm.TryAdvanceCommanderStoreupAwardIndexTx(tx, commanderID, storeupID, awardIndex)
		if err != nil {
			return err
		}
		if !advanced {
			return sentinelWrongTier
		}

		switch dropType {
		case consts.DROP_TYPE_RESOURCE:
			if err := client.Commander.AddResourceTx(tx, dropID, dropCount); err != nil {
				return err
			}
		case consts.DROP_TYPE_ITEM:
			if err := client.Commander.AddItemTx(tx, dropID, dropCount); err != nil {
				return err
			}
		case consts.DROP_TYPE_EQUIP:
			if err := client.Commander.AddOwnedEquipmentTx(tx, dropID, dropCount); err != nil {
				return err
			}
		case consts.DROP_TYPE_SHIP:
			for i := uint32(0); i < dropCount; i++ {
				if _, err := client.Commander.AddShipTx(tx, dropID); err != nil {
					return err
				}
			}
		case consts.DROP_TYPE_FURNITURE:
			if err := orm.AddCommanderFurnitureTx(tx, commanderID, dropID, dropCount, now); err != nil {
				return err
			}
		case consts.DROP_TYPE_SKIN:
			for i := uint32(0); i < dropCount; i++ {
				if err := client.Commander.GiveSkinTx(tx, dropID); err != nil {
					return err
				}
			}
		default:
			return sentinelUnsupported
		}

		return nil
	})
	if err != nil {
		switch {
		case errors.Is(err, sentinelWrongTier):
			response.Result = proto.Uint32(resultWrongTier)
		case errors.Is(err, sentinelNotEligible):
			response.Result = proto.Uint32(resultNotEligible)
		case errors.Is(err, sentinelUnsupported):
			response.Result = proto.Uint32(resultUnsupported)
		default:
			response.Result = proto.Uint32(resultDBError)
		}
		return client.SendMessage(17006, &response)
	}

	return client.SendMessage(17006, &response)
}

func storeupStarCount(commanderID uint32, groups []uint32) (uint32, error) {
	var rows []storeupGroupStar
	if err := orm.GormDB.Raw(`
	SELECT
		ship_id / 10 AS group_id,
		MAX(ships.star) AS max_star
	FROM owned_ships
	INNER JOIN ships ON owned_ships.ship_id = ships.template_id
	WHERE owner_id = ?
	GROUP BY group_id
	`, commanderID).Scan(&rows).Error; err != nil {
		return 0, err
	}

	lookup := make(map[uint32]uint32, len(rows))
	for i := range rows {
		lookup[rows[i].GroupID] = rows[i].MaxStar
	}

	var total uint32
	for _, groupID := range groups {
		if star, ok := lookup[groupID]; ok {
			total += star
		}
	}
	return total, nil
}

func loadStoreupDataTemplate(id uint32) (*storeupDataTemplate, bool, error) {
	key := fmt.Sprintf("%d", id)
	if entry, err := orm.GetConfigEntry(orm.GormDB, storeupDataTemplateCategory, key); err == nil {
		var out storeupDataTemplate
		if err := json.Unmarshal(entry.Data, &out); err != nil {
			return nil, false, err
		}
		if out.ID == 0 {
			out.ID = id
		}
		return &out, true, nil
	}

	entries, err := orm.ListConfigEntries(orm.GormDB, storeupDataTemplateCategory)
	if err != nil {
		return nil, false, err
	}
	for i := range entries {
		var single storeupDataTemplate
		if err := json.Unmarshal(entries[i].Data, &single); err == nil {
			if single.ID == id {
				return &single, true, nil
			}
		}
		var list []storeupDataTemplate
		if err := json.Unmarshal(entries[i].Data, &list); err == nil {
			for j := range list {
				if list[j].ID == id {
					return &list[j], true, nil
				}
			}
		}
	}
	return nil, false, nil
}
