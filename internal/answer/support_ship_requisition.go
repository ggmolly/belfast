package answer

import (
	"context"
	"errors"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/consts"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"github.com/ggmolly/belfast/internal/rng"
	"github.com/jackc/pgx/v5"
	"google.golang.org/protobuf/proto"
)

const (
	supportRequisitionItemID = 15001

	supportRequisitionResultOK              = 0
	supportRequisitionResultFailed          = 1
	supportRequisitionResultNotEnoughMedals = 2
	supportRequisitionResultLimitReached    = 30
)

var supportRequisitionRng = rng.NewLockedRand()

func SupportShipRequisition(buffer *[]byte, client *connection.Client) (int, int, error) {
	var data protobuf.CS_16100
	if err := proto.Unmarshal(*buffer, &data); err != nil {
		return 0, 16100, err
	}

	response := protobuf.SC_16101{Result: proto.Uint32(supportRequisitionResultFailed)}
	count := data.GetCnt()
	if count == 0 || count > consts.MaxBuildWorkCount {
		return client.SendMessage(16101, &response)
	}

	config, err := orm.LoadSupportRequisitionConfig()
	if err != nil {
		return client.SendMessage(16101, &response)
	}

	now := time.Now().UTC()
	monthReset := client.Commander.EnsureSupportRequisitionMonth(now)
	_ = monthReset
	if client.Commander.SupportRequisitionCount+count > config.MonthlyCap {
		response.Result = proto.Uint32(supportRequisitionResultLimitReached)
		return client.SendMessage(16101, &response)
	}

	cost := config.Cost * count
	if !client.Commander.HasEnoughItem(supportRequisitionItemID, cost) {
		response.Result = proto.Uint32(supportRequisitionResultNotEnoughMedals)
		return client.SendMessage(16101, &response)
	}

	shipTemplates := make([]uint32, 0, count)
	for i := uint32(0); i < count; i++ {
		rarity, err := selectSupportRequisitionRarity(config.RarityWeights)
		if err != nil {
			return client.SendMessage(16101, &response)
		}
		ship, err := orm.GetRandomRequisitionShipByRarity(rarity)
		if err != nil {
			return client.SendMessage(16101, &response)
		}
		shipTemplates = append(shipTemplates, ship.TemplateID)
	}

	ctx := context.Background()
	ships := make([]*orm.OwnedShip, 0, count)
	shipIDs := make([]uint32, 0, count)
	err = orm.WithPGXTx(ctx, func(tx pgx.Tx) error {
		if err := client.Commander.ConsumeItemTx(ctx, tx, supportRequisitionItemID, cost); err != nil {
			return err
		}
		client.Commander.SupportRequisitionCount += count
		if err := client.Commander.SaveTx(ctx, tx); err != nil {
			return err
		}

		for _, shipID := range shipTemplates {
			owned, err := client.Commander.AddShipTx(ctx, tx, shipID)
			if err != nil {
				return err
			}
			ships = append(ships, owned)
			shipIDs = append(shipIDs, owned.ID)
		}
		return nil
	})
	if err != nil {
		if err.Error() == "not enough items" {
			response.Result = proto.Uint32(supportRequisitionResultNotEnoughMedals)
			return client.SendMessage(16101, &response)
		}
		response.Result = proto.Uint32(supportRequisitionResultFailed)
		return client.SendMessage(16101, &response)
	}

	flags, err := orm.ListRandomFlagShipPhantoms(client.Commander.CommanderID, shipIDs)
	if err != nil {
		response.Result = proto.Uint32(supportRequisitionResultFailed)
		return client.SendMessage(16101, &response)
	}
	shadows, err := orm.ListOwnedShipShadowSkins(client.Commander.CommanderID, shipIDs)
	if err != nil {
		response.Result = proto.Uint32(supportRequisitionResultFailed)
		return client.SendMessage(16101, &response)
	}

	response.Result = proto.Uint32(supportRequisitionResultOK)
	response.ShipList = make([]*protobuf.SHIPINFO, len(ships))
	for i, ship := range ships {
		response.ShipList[i] = orm.ToProtoOwnedShip(*ship, flags[ship.ID], shadows[ship.ID])
	}

	return client.SendMessage(16101, &response)
}

func selectSupportRequisitionRarity(weights []orm.SupportRarityWeight) (uint32, error) {
	var total uint32
	for _, entry := range weights {
		total += entry.Weight
	}
	if total == 0 {
		return 0, errors.New("support requisition weights are empty")
	}
	roll := supportRequisitionRng.Uint32N(total) + 1
	var cumulative uint32
	for _, entry := range weights {
		cumulative += entry.Weight
		if roll <= cumulative {
			return entry.Rarity, nil
		}
	}
	return 0, errors.New("support requisition rarity selection failed")
}
