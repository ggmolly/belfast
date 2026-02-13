package answer

import (
	"context"
	"encoding/json"
	"errors"
	"math"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"github.com/jackc/pgx/v5"
	"google.golang.org/protobuf/proto"
)

const (
	destroyEquipmentsResultOK             uint32 = 0
	destroyEquipmentsResultGenericFailure uint32 = 1
	destroyEquipmentsResultNotEnough      uint32 = 2
	destroyEquipmentsResultUnknownEquip   uint32 = 3
)

func DestroyEquipments(buffer *[]byte, client *connection.Client) (int, int, error) {
	response := protobuf.SC_14009{Result: proto.Uint32(destroyEquipmentsResultGenericFailure)}

	var payload protobuf.CS_14008
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return client.SendMessage(14009, &response)
	}
	if len(payload.GetEquipList()) == 0 {
		return client.SendMessage(14009, &response)
	}

	// Many handlers assume the commander is already loaded from the login flow,
	// but tests and dev tooling may call packet handlers directly.
	if client.Commander.OwnedEquipmentMap == nil || client.Commander.OwnedResourcesMap == nil || client.Commander.CommanderItemsMap == nil || client.Commander.MiscItemsMap == nil {
		if err := client.Commander.Load(); err != nil {
			return client.SendMessage(14009, &response)
		}
	}

	equipmentCounts := make(map[uint32]uint32)
	for _, entry := range payload.GetEquipList() {
		equipmentID := entry.GetId()
		count := entry.GetCount()
		if equipmentID == 0 || count == 0 {
			return client.SendMessage(14009, &response)
		}
		next := uint64(equipmentCounts[equipmentID]) + uint64(count)
		if next > math.MaxUint32 {
			return client.SendMessage(14009, &response)
		}
		equipmentCounts[equipmentID] = uint32(next)
	}

	totalGold := uint64(0)
	items := make(map[uint32]uint64)
	ctx := context.Background()
	for equipmentID, count := range equipmentCounts {
		owned := client.Commander.GetOwnedEquipment(equipmentID)
		if owned == nil || owned.Count < count {
			response.Result = proto.Uint32(destroyEquipmentsResultNotEnough)
			return client.SendMessage(14009, &response)
		}

		var destroyGold int64
		var destroyItemsRaw []byte
		err := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT destroy_gold, destroy_item
FROM equipments
WHERE id = $1
`, int64(equipmentID)).Scan(&destroyGold, &destroyItemsRaw)
		err = db.MapNotFound(err)
		if err != nil {
			if db.IsNotFound(err) {
				response.Result = proto.Uint32(destroyEquipmentsResultUnknownEquip)
				return client.SendMessage(14009, &response)
			}
			return client.SendMessage(14009, &response)
		}

		totalGold += uint64(destroyGold) * uint64(count)
		if totalGold > math.MaxUint32 {
			return client.SendMessage(14009, &response)
		}

		rewards, err := parseDestroyEquipmentItems(destroyItemsRaw)
		if err != nil {
			return client.SendMessage(14009, &response)
		}
		for itemID, per := range rewards {
			grant := uint64(per) * uint64(count)
			items[itemID] += grant
			if items[itemID] > math.MaxUint32 {
				return client.SendMessage(14009, &response)
			}
		}
	}

	err := orm.WithPGXTx(ctx, func(tx pgx.Tx) error {
		for equipmentID, count := range equipmentCounts {
			if err := client.Commander.RemoveOwnedEquipmentTx(ctx, tx, equipmentID, count); err != nil {
				response.Result = proto.Uint32(destroyEquipmentsResultNotEnough)
				return err
			}
		}
		if totalGold != 0 {
			if err := client.Commander.AddResourceTx(ctx, tx, 1, uint32(totalGold)); err != nil {
				return err
			}
		}
		for itemID, count := range items {
			if count == 0 {
				continue
			}
			if err := client.Commander.AddItemTx(ctx, tx, itemID, uint32(count)); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return client.SendMessage(14009, &response)
	}

	response.Result = proto.Uint32(destroyEquipmentsResultOK)
	return client.SendMessage(14009, &response)
}

func parseDestroyEquipmentItems(raw json.RawMessage) (map[uint32]uint32, error) {
	if len(raw) == 0 {
		return map[uint32]uint32{}, nil
	}
	var pairs [][]uint32
	if err := json.Unmarshal(raw, &pairs); err != nil {
		return nil, err
	}

	out := make(map[uint32]uint32, len(pairs))
	for _, pair := range pairs {
		if len(pair) != 2 {
			return nil, errors.New("invalid destory_item")
		}
		itemID := pair[0]
		count := pair[1]
		if itemID == 0 || count == 0 {
			continue
		}
		next := uint64(out[itemID]) + uint64(count)
		if next > math.MaxUint32 {
			return nil, errors.New("destory_item overflow")
		}
		out[itemID] = uint32(next)
	}
	return out, nil
}
