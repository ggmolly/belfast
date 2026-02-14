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

const (
	upgradeEquipmentOnShipResultOK             uint32 = 0
	upgradeEquipmentOnShipResultGenericFailure uint32 = 1
)

func UpgradeEquipmentOnShip14002(buffer *[]byte, client *connection.Client) (int, int, error) {
	response := protobuf.SC_14003{Result: proto.Uint32(upgradeEquipmentOnShipResultGenericFailure)}

	if client == nil {
		return 0, 14002, errors.New("nil client")
	}
	if client.Commander == nil {
		return client.SendMessage(14003, &response)
	}

	var payload protobuf.CS_14002
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return client.SendMessage(14003, &response)
	}

	shipID := payload.GetShipId()
	pos := payload.GetPos()
	lv := payload.GetLv()
	if shipID == 0 || pos == 0 || lv == 0 {
		return client.SendMessage(14003, &response)
	}

	if client.Commander.OwnedShipsMap == nil || client.Commander.OwnedResourcesMap == nil || client.Commander.CommanderItemsMap == nil || client.Commander.MiscItemsMap == nil {
		if err := client.Commander.Load(); err != nil {
			return client.SendMessage(14003, &response)
		}
	}

	ship, ok := client.Commander.OwnedShipsMap[shipID]
	if !ok {
		return client.SendMessage(14003, &response)
	}

	current, err := orm.GetOwnedShipEquipment(client.Commander.CommanderID, ship.ID, pos)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			current = buildShipEquipmentFromMemory(client.Commander.CommanderID, ship, pos)
		} else {
			return 0, 14002, err
		}
	}
	if current.EquipID == 0 {
		return client.SendMessage(14003, &response)
	}

	upgradedID, itemCosts, coinCost, ok := computeEquipmentUpgradeCosts(current.EquipID, lv)
	if !ok {
		return client.SendMessage(14003, &response)
	}
	if coinCost != 0 && !client.Commander.HasEnoughResource(1, coinCost) {
		return client.SendMessage(14003, &response)
	}
	for itemID, count := range itemCosts {
		if !client.Commander.HasEnoughItem(itemID, count) {
			return client.SendMessage(14003, &response)
		}
	}

	ctx := context.Background()
	err = orm.WithPGXTx(ctx, func(tx pgx.Tx) error {
		if coinCost != 0 {
			if err := client.Commander.ConsumeResourceTx(ctx, tx, 1, coinCost); err != nil {
				return err
			}
		}
		for itemID, count := range itemCosts {
			if count == 0 {
				continue
			}
			if err := client.Commander.ConsumeItemTx(ctx, tx, itemID, count); err != nil {
				return err
			}
		}

		current.EquipID = upgradedID
		return orm.UpsertOwnedShipEquipmentTx(ctx, tx, current)
	})
	if err != nil {
		return client.SendMessage(14003, &response)
	}

	applyShipEquipmentUpdate(ship, current)
	response.Result = proto.Uint32(upgradeEquipmentOnShipResultOK)
	return client.SendMessage(14003, &response)
}
