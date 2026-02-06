package answer

import (
	"errors"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/consts"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"github.com/ggmolly/belfast/internal/rng"
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

	config, err := orm.LoadSupportRequisitionConfig(orm.GormDB)
	if err != nil {
		return client.SendMessage(16101, &response)
	}

	now := time.Now().UTC()
	monthReset := client.Commander.EnsureSupportRequisitionMonth(now)
	if monthReset {
		if err := orm.GormDB.Save(client.Commander).Error; err != nil {
			return client.SendMessage(16101, &response)
		}
	}
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

	tx := orm.GormDB.Begin()
	if tx.Error != nil {
		return client.SendMessage(16101, &response)
	}
	if err := client.Commander.ConsumeItemTx(tx, supportRequisitionItemID, cost); err != nil {
		tx.Rollback()
		response.Result = proto.Uint32(supportRequisitionResultNotEnoughMedals)
		return client.SendMessage(16101, &response)
	}
	client.Commander.SupportRequisitionCount += count
	if err := client.Commander.SaveTx(tx); err != nil {
		tx.Rollback()
		response.Result = proto.Uint32(supportRequisitionResultFailed)
		return client.SendMessage(16101, &response)
	}

	ships := make([]*orm.OwnedShip, 0, count)
	shipIDs := make([]uint32, 0, count)
	for _, shipID := range shipTemplates {
		owned, err := client.Commander.AddShipTx(tx, shipID)
		if err != nil {
			tx.Rollback()
			response.Result = proto.Uint32(supportRequisitionResultFailed)
			return client.SendMessage(16101, &response)
		}
		ships = append(ships, owned)
		shipIDs = append(shipIDs, owned.ID)
	}
	if err := tx.Commit().Error; err != nil {
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
