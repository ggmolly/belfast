package answer

import (
	"context"
	"errors"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"github.com/jackc/pgx/v5"
	"google.golang.org/protobuf/proto"
)

const (
	upgradeEquipmentInBagResultOK             uint32 = 0
	upgradeEquipmentInBagResultGenericFailure uint32 = 1
)

func UpgradeEquipmentInBag14004(buffer *[]byte, client *connection.Client) (int, int, error) {
	response := protobuf.SC_14005{Result: proto.Uint32(upgradeEquipmentInBagResultGenericFailure)}

	if client == nil {
		return 0, 14004, errors.New("nil client")
	}
	if client.Commander == nil {
		return client.SendMessage(14005, &response)
	}

	var payload protobuf.CS_14004
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return client.SendMessage(14005, &response)
	}

	equipID := payload.GetEquipId()
	lv := payload.GetLv()
	if equipID == 0 || lv == 0 {
		return client.SendMessage(14005, &response)
	}

	if client.Commander.OwnedEquipmentMap == nil || client.Commander.OwnedResourcesMap == nil || client.Commander.CommanderItemsMap == nil || client.Commander.MiscItemsMap == nil {
		if err := client.Commander.Load(); err != nil {
			return client.SendMessage(14005, &response)
		}
	}

	owned := client.Commander.GetOwnedEquipment(equipID)
	if owned == nil || owned.Count < 1 {
		return client.SendMessage(14005, &response)
	}

	upgradedID, itemCosts, coinCost, ok := computeEquipmentUpgradeCosts(equipID, lv)
	if !ok {
		return client.SendMessage(14005, &response)
	}
	if coinCost != 0 && !client.Commander.HasEnoughResource(1, coinCost) {
		return client.SendMessage(14005, &response)
	}
	for itemID, count := range itemCosts {
		if !client.Commander.HasEnoughItem(itemID, count) {
			return client.SendMessage(14005, &response)
		}
	}

	ctx := context.Background()
	if err := orm.WithPGXTx(ctx, func(tx pgx.Tx) error {
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
		if err := client.Commander.RemoveOwnedEquipmentTx(ctx, tx, equipID, 1); err != nil {
			return err
		}
		return client.Commander.AddOwnedEquipmentTx(ctx, tx, upgradedID, 1)
	}); err != nil {
		return client.SendMessage(14005, &response)
	}

	response.Result = proto.Uint32(upgradeEquipmentInBagResultOK)
	return client.SendMessage(14005, &response)
}

func computeEquipmentUpgradeCosts(startID uint32, lv uint32) (uint32, map[uint32]uint32, uint32, bool) {
	currentID := startID
	itemCosts := make(map[uint32]uint32)
	var coinCost uint32
	for i := uint32(0); i < lv; i++ {
		current, err := loadEquipmentConfig(currentID)
		if err != nil {
			return 0, nil, 0, false
		}
		if current.Next == 0 {
			return 0, nil, 0, false
		}
		coinCost += current.TransUseGold
		if err := addTransUseItems(itemCosts, current.TransUseItem); err != nil {
			return 0, nil, 0, false
		}
		currentID = uint32(current.Next)
	}
	return currentID, itemCosts, coinCost, true
}
