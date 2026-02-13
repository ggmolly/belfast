package answer

import (
	"context"
	"errors"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"github.com/jackc/pgx/v5"
	"google.golang.org/protobuf/proto"
)

func TransformEquipmentInBag14015(buffer *[]byte, client *connection.Client) (int, int, error) {
	var data protobuf.CS_14015
	if err := proto.Unmarshal(*buffer, &data); err != nil {
		return 0, 14015, err
	}

	response := protobuf.SC_14016{Result: proto.Uint32(0)}
	equipID := data.GetEquipId()
	if equipID == 0 {
		response.Result = proto.Uint32(1)
		return client.SendMessage(14016, &response)
	}
	owned := client.Commander.GetOwnedEquipment(equipID)
	if owned == nil || owned.Count == 0 {
		response.Result = proto.Uint32(1)
		return client.SendMessage(14016, &response)
	}

	upgrade, err := orm.GetEquipUpgradeDataTx(data.GetUpgradeId())
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			response.Result = proto.Uint32(1)
			return client.SendMessage(14016, &response)
		}
		return 0, 14015, err
	}
	if upgrade.UpgradeFrom != equipID {
		response.Result = proto.Uint32(1)
		return client.SendMessage(14016, &response)
	}
	targetEquipID := upgrade.TargetID
	if targetEquipID == 0 {
		response.Result = proto.Uint32(1)
		return client.SendMessage(14016, &response)
	}

	if client.Commander.GetResourceCount(1) < upgrade.CoinConsume {
		response.Result = proto.Uint32(1)
		return client.SendMessage(14016, &response)
	}
	for _, cost := range upgrade.MaterialCost {
		if client.Commander.GetItemCount(cost.ItemID) < cost.Count {
			response.Result = proto.Uint32(1)
			return client.SendMessage(14016, &response)
		}
	}

	ctx := context.Background()
	err = orm.WithPGXTx(ctx, func(tx pgx.Tx) error {
		if upgrade.CoinConsume != 0 {
			if err := client.Commander.ConsumeResourceTx(ctx, tx, 1, upgrade.CoinConsume); err != nil {
				return err
			}
		}
		for _, cost := range upgrade.MaterialCost {
			if err := client.Commander.ConsumeItemTx(ctx, tx, cost.ItemID, cost.Count); err != nil {
				return err
			}
		}
		if err := client.Commander.RemoveOwnedEquipmentTx(ctx, tx, equipID, 1); err != nil {
			return err
		}
		return client.Commander.AddOwnedEquipmentTx(ctx, tx, targetEquipID, 1)
	})
	if err != nil {
		response.Result = proto.Uint32(1)
		return 0, 14015, err
	}
	return client.SendMessage(14016, &response)
}
