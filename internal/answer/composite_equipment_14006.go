package answer

import (
	"fmt"
	"math"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/logger"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func CompositeEquipment(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_14006
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 14006, err
	}

	response := protobuf.SC_14007{Result: proto.Uint32(1)}
	composeID := payload.GetId()
	num := payload.GetNum()
	if composeID == 0 || num == 0 {
		return client.SendMessage(14007, &response)
	}

	recipe, err := orm.GetComposeDataTemplateEntry(orm.GormDB, composeID)
	if err != nil {
		return 0, 14006, err
	}
	if recipe == nil {
		logger.Scope("Packets/Equip").With(
			logger.Field{Key: "compose_id", Value: fmt.Sprintf("%d", composeID)},
		).Warn("compose recipe missing")
		return client.SendMessage(14007, &response)
	}
	if recipe.EquipID == 0 || recipe.MaterialID == 0 || recipe.MaterialNum == 0 {
		return client.SendMessage(14007, &response)
	}

	if client.Commander.EquipmentBagCount()+num > equipBagMax {
		return client.SendMessage(14007, &response)
	}

	goldCost64 := uint64(recipe.GoldNum) * uint64(num)
	materialCost64 := uint64(recipe.MaterialNum) * uint64(num)
	if goldCost64 > math.MaxUint32 || materialCost64 > math.MaxUint32 {
		return client.SendMessage(14007, &response)
	}
	goldCost := uint32(goldCost64)
	materialCost := uint32(materialCost64)

	if client.Commander.GetResourceCount(1) < goldCost {
		return client.SendMessage(14007, &response)
	}
	if client.Commander.GetItemCount(recipe.MaterialID) < materialCost {
		return client.SendMessage(14007, &response)
	}

	tx := orm.GormDB.Begin()
	if tx.Error != nil {
		return client.SendMessage(14007, &response)
	}
	if goldCost != 0 {
		if err := client.Commander.ConsumeResourceTx(tx, 1, goldCost); err != nil {
			tx.Rollback()
			return client.SendMessage(14007, &response)
		}
	}
	if materialCost != 0 {
		if err := client.Commander.ConsumeItemTx(tx, recipe.MaterialID, materialCost); err != nil {
			tx.Rollback()
			return client.SendMessage(14007, &response)
		}
	}
	if err := client.Commander.AddOwnedEquipmentTx(tx, recipe.EquipID, num); err != nil {
		tx.Rollback()
		return client.SendMessage(14007, &response)
	}
	if err := tx.Commit().Error; err != nil {
		return 0, 14006, err
	}

	response.Result = proto.Uint32(0)
	return client.SendMessage(14007, &response)
}
