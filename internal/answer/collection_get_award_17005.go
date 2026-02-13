package answer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/consts"
	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"github.com/jackc/pgx/v5"
	"google.golang.org/protobuf/proto"
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
	ctx := context.Background()
	err = db.DefaultStore.WithPGXTx(ctx, func(tx pgx.Tx) error {
		if template.Level[awardIndex-1] > starCount {
			return sentinelNotEligible
		}
		advanced, err := orm.TryAdvanceCommanderStoreupAwardIndexTx(ctx, tx, commanderID, storeupID, awardIndex)
		if err != nil {
			return err
		}
		if !advanced {
			return sentinelWrongTier
		}

		switch dropType {
		case consts.DROP_TYPE_RESOURCE:
			if err := client.Commander.AddResourceTx(ctx, tx, dropID, dropCount); err != nil {
				return err
			}
		case consts.DROP_TYPE_ITEM:
			if err := client.Commander.AddItemTx(ctx, tx, dropID, dropCount); err != nil {
				return err
			}
		case consts.DROP_TYPE_EQUIP:
			if err := addOwnedEquipmentPGXTx(ctx, tx, client.Commander, dropID, dropCount); err != nil {
				return err
			}
		case consts.DROP_TYPE_SHIP:
			for i := uint32(0); i < dropCount; i++ {
				if _, err := client.Commander.AddShipTx(ctx, tx, dropID); err != nil {
					return err
				}
			}
		case consts.DROP_TYPE_FURNITURE:
			if err := orm.AddCommanderFurnitureTx(ctx, tx, commanderID, dropID, dropCount, now); err != nil {
				return err
			}
		case consts.DROP_TYPE_SKIN:
			for i := uint32(0); i < dropCount; i++ {
				if err := client.Commander.GiveSkinTx(ctx, tx, dropID); err != nil {
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
	rows, err := db.DefaultStore.Pool.Query(context.Background(), `
SELECT
	owned_ships.ship_id / 10 AS group_id,
	MAX(ships.star) AS max_star
FROM owned_ships
INNER JOIN ships ON owned_ships.ship_id = ships.template_id
WHERE owner_id = $1
GROUP BY group_id
`, int64(commanderID))
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	groupsWithStars := make([]storeupGroupStar, 0)
	for rows.Next() {
		var row storeupGroupStar
		if err := rows.Scan(&row.GroupID, &row.MaxStar); err != nil {
			return 0, err
		}
		groupsWithStars = append(groupsWithStars, row)
	}

	lookup := make(map[uint32]uint32, len(groupsWithStars))
	for i := range groupsWithStars {
		lookup[groupsWithStars[i].GroupID] = groupsWithStars[i].MaxStar
	}

	var total uint32
	for _, groupID := range groups {
		if star, ok := lookup[groupID]; ok {
			total += star
		}
	}
	return total, nil
}

func addOwnedEquipmentPGXTx(ctx context.Context, tx pgx.Tx, commander *orm.Commander, equipmentID uint32, count uint32) error {
	if count == 0 {
		return nil
	}
	if commander.OwnedEquipmentMap == nil {
		commander.RebuildOwnedEquipmentMap()
	}
	_, err := tx.Exec(ctx, `
INSERT INTO owned_equipments (commander_id, equipment_id, count)
VALUES ($1, $2, $3)
ON CONFLICT (commander_id, equipment_id)
DO UPDATE SET count = owned_equipments.count + EXCLUDED.count
`, int64(commander.CommanderID), int64(equipmentID), int64(count))
	if err != nil {
		return err
	}
	if existing, ok := commander.OwnedEquipmentMap[equipmentID]; ok {
		existing.Count += count
		return nil
	}
	commander.OwnedEquipments = append(commander.OwnedEquipments, orm.OwnedEquipment{CommanderID: commander.CommanderID, EquipmentID: equipmentID, Count: count})
	commander.RebuildOwnedEquipmentMap()
	return nil
}

func loadStoreupDataTemplate(id uint32) (*storeupDataTemplate, bool, error) {
	key := fmt.Sprintf("%d", id)
	if entry, err := orm.GetConfigEntry(storeupDataTemplateCategory, key); err == nil {
		var out storeupDataTemplate
		if err := json.Unmarshal(entry.Data, &out); err != nil {
			return nil, false, err
		}
		if out.ID == 0 {
			out.ID = id
		}
		return &out, true, nil
	}

	entries, err := orm.ListConfigEntries(storeupDataTemplateCategory)
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
