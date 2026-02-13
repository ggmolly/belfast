package answer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"github.com/jackc/pgx/v5"
	"google.golang.org/protobuf/proto"
)

const revertEquipmentItemID uint32 = 15007

func RevertEquipment(buffer *[]byte, client *connection.Client) (int, int, error) {
	var data protobuf.CS_14010
	if err := proto.Unmarshal(*buffer, &data); err != nil {
		return 0, 14010, err
	}

	response := protobuf.SC_14011{Result: proto.Uint32(0)}
	equipID := data.GetEquipId()
	if equipID == 0 {
		response.Result = proto.Uint32(1)
		return client.SendMessage(14011, &response)
	}
	owned := client.Commander.GetOwnedEquipment(equipID)
	if owned == nil || owned.Count == 0 {
		response.Result = proto.Uint32(1)
		return client.SendMessage(14011, &response)
	}
	if !client.Commander.HasEnoughItem(revertEquipmentItemID, 1) {
		response.Result = proto.Uint32(1)
		return client.SendMessage(14011, &response)
	}

	rootEquipID, refundItems, refundCoins, ok, err := computeRevertEquipmentRefunds(equipID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			response.Result = proto.Uint32(1)
			return client.SendMessage(14011, &response)
		}
		return 0, 14010, err
	}
	if !ok {
		response.Result = proto.Uint32(1)
		return client.SendMessage(14011, &response)
	}

	ctx := context.Background()
	err = orm.WithPGXTx(ctx, func(tx pgx.Tx) error {
		if err := client.Commander.ConsumeItemTx(ctx, tx, revertEquipmentItemID, 1); err != nil {
			return err
		}
		if err := client.Commander.RemoveOwnedEquipmentTx(ctx, tx, equipID, 1); err != nil {
			return err
		}
		if err := client.Commander.AddOwnedEquipmentTx(ctx, tx, rootEquipID, 1); err != nil {
			return err
		}
		for itemID, count := range refundItems {
			if err := client.Commander.AddItemTx(ctx, tx, itemID, count); err != nil {
				return err
			}
		}
		if refundCoins != 0 {
			if err := client.Commander.AddResourceTx(ctx, tx, 1, refundCoins); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		response.Result = proto.Uint32(1)
		if errors.Is(err, db.ErrNotFound) {
			return client.SendMessage(14011, &response)
		}
		return 0, 14010, err
	}
	return client.SendMessage(14011, &response)
}

func computeRevertEquipmentRefunds(equipID uint32) (uint32, map[uint32]uint32, uint32, bool, error) {
	current, err := loadEquipmentConfig(equipID)
	if err != nil {
		return 0, nil, 0, false, err
	}
	if current.Prev == 0 || current.Level <= 1 {
		return 0, nil, 0, false, nil
	}

	refundItems := make(map[uint32]uint32)
	var refundCoins uint32
	for current.Prev != 0 {
		prevID := uint32(current.Prev)
		prev, err := loadEquipmentConfig(prevID)
		if err != nil {
			return 0, nil, 0, false, err
		}
		refundCoins += prev.TransUseGold
		if err := addTransUseItems(refundItems, prev.TransUseItem); err != nil {
			return 0, nil, 0, false, err
		}
		current = prev
	}

	return current.ID, refundItems, refundCoins, true, nil
}

func loadEquipmentConfig(equipID uint32) (*orm.Equipment, error) {
	entry, err := orm.GetConfigEntry("sharecfgdata/equip_data_statistics.json", fmt.Sprintf("%d", equipID))
	if err != nil {
		return nil, err
	}
	var equipment orm.Equipment
	if err := json.Unmarshal(entry.Data, &equipment); err != nil {
		return nil, err
	}
	return &equipment, nil
}
