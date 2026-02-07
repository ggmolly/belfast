package answer

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

type storedChild struct {
	Id string `json:"id"`
	X  uint32 `json:"x"`
	Y  uint32 `json:"y"`
}

type storedFurniturePut struct {
	Id     string        `json:"id"`
	X      uint32        `json:"x"`
	Y      uint32        `json:"y"`
	Dir    uint32        `json:"dir"`
	Child  []storedChild `json:"child"`
	Parent uint64        `json:"parent"`
	ShipId uint32        `json:"shipId"`
}

func AddDormShip19002(buffer *[]byte, client *connection.Client) (int, int, error) {
	var request protobuf.CS_19002
	if err := proto.Unmarshal(*buffer, &request); err != nil {
		return 0, 19003, err
	}
	commanderID := client.Commander.CommanderID
	shipID := request.GetShipId()
	shipType := request.GetType()

	tx := orm.GormDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	state, err := orm.GetOrCreateCommanderDormStateTx(tx, commanderID)
	if err != nil {
		tx.Rollback()
		return 0, 19003, err
	}

	var ship orm.OwnedShip
	if err := tx.Where("owner_id = ? AND id = ?", commanderID, shipID).First(&ship).Error; err != nil {
		tx.Rollback()
		return 0, 19003, err
	}

	now := uint32(time.Now().Unix())
	if shipType == 1 {
		// Train
		ship.State = 5
		ship.StateInfo1 = now
		ship.StateInfo2 = 0
		if state.NextTimestamp == 0 {
			state.NextTimestamp = now + 15
			state.LoadTime = now
		}
	} else if shipType == 2 {
		// Rest
		ship.State = 2
	} else {
		// Unsupported type
		shipType = 0
	}
	if err := tx.Save(&ship).Error; err != nil {
		tx.Rollback()
		return 0, 19003, err
	}
	if err := tx.Save(state).Error; err != nil {
		tx.Rollback()
		return 0, 19003, err
	}
	if err := tx.Commit().Error; err != nil {
		return 0, 19003, err
	}

	response := protobuf.SC_19003{Result: proto.Uint32(0)}
	return client.SendMessage(19003, &response)
}

func ExitDormShip19004(buffer *[]byte, client *connection.Client) (int, int, error) {
	var request protobuf.CS_19004
	if err := proto.Unmarshal(*buffer, &request); err != nil {
		return 0, 19005, err
	}
	commanderID := client.Commander.CommanderID
	shipID := request.GetShipId()

	tx := orm.GormDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	if _, err := tickDormStateTx(tx, commanderID, uint32(time.Now().Unix())); err != nil {
		tx.Rollback()
		return 0, 19005, err
	}
	var ship orm.OwnedShip
	if err := tx.Preload("Ship").Where("owner_id = ? AND id = ?", commanderID, shipID).First(&ship).Error; err != nil {
		tx.Rollback()
		return 0, 19005, err
	}
	gained := ship.StateInfo2
	_ = applyOwnedShipExpGain(&ship, gained)
	ship.State = 0
	ship.StateInfo1 = 0
	ship.StateInfo2 = 0
	ship.StateInfo3 = 0
	ship.StateInfo4 = 0
	if err := tx.Save(&ship).Error; err != nil {
		tx.Rollback()
		return 0, 19005, err
	}
	if err := tx.Commit().Error; err != nil {
		return 0, 19005, err
	}
	response := protobuf.SC_19005{Result: proto.Uint32(0), Exp: proto.Uint32(gained)}
	return client.SendMessage(19005, &response)
}

func BuyFurniture19006(buffer *[]byte, client *connection.Client) (int, int, error) {
	var request protobuf.CS_19006
	if err := proto.Unmarshal(*buffer, &request); err != nil {
		return 0, 19007, err
	}
	commanderID := client.Commander.CommanderID
	currency := request.GetCurrency()
	ids := request.GetFurnitureId()
	if len(ids) == 0 {
		resp := protobuf.SC_19007{Result: proto.Uint32(0)}
		return client.SendMessage(19007, &resp)
	}

	tx := orm.GormDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// Verify cost.
	var totalCost uint32
	for _, furnitureID := range ids {
		shop, err := orm.GetConfigEntry(tx, "ShareCfg/furniture_shop_template.json", fmt.Sprintf("%d", furnitureID))
		if err != nil {
			tx.Rollback()
			return 0, 19007, err
		}
		var entry struct {
			GemPrice      uint32 `json:"gem_price"`
			DormIconPrice uint32 `json:"dorm_icon_price"`
		}
		if err := json.Unmarshal(shop.Data, &entry); err != nil {
			tx.Rollback()
			return 0, 19007, err
		}
		var cost uint32
		switch currency {
		case 4:
			cost = entry.GemPrice
		case 6:
			cost = entry.DormIconPrice
		default:
			tx.Rollback()
			return 0, 19007, fmt.Errorf("unsupported currency %d", currency)
		}
		if cost == 0 {
			tx.Rollback()
			return 0, 19007, fmt.Errorf("furniture %d not purchasable with currency %d", furnitureID, currency)
		}
		totalCost += cost
	}

	// Deduct currency.
	if err := client.Commander.ConsumeResourceTx(tx, currency, totalCost); err != nil {
		tx.Rollback()
		resp := protobuf.SC_19007{Result: proto.Uint32(1)}
		return client.SendMessage(19007, &resp)
	}

	now := uint32(time.Now().Unix())
	for _, furnitureID := range ids {
		if err := orm.AddCommanderFurnitureTx(tx, commanderID, furnitureID, 1, now); err != nil {
			tx.Rollback()
			return 0, 19007, err
		}
	}
	if err := tx.Commit().Error; err != nil {
		return 0, 19007, err
	}
	resp := protobuf.SC_19007{Result: proto.Uint32(0)}
	return client.SendMessage(19007, &resp)
}

func PutFurniture19008(buffer *[]byte, client *connection.Client) (int, int, error) {
	var request protobuf.CS_19008
	if err := proto.Unmarshal(*buffer, &request); err != nil {
		return 0, 19009, err
	}
	commanderID := client.Commander.CommanderID
	floor := request.GetFloor()
	state, err := orm.GetOrCreateCommanderDormState(commanderID)
	if err != nil {
		return 0, 19009, err
	}
	const maxDormFloor = uint32(3)
	if floor == 0 || floor > maxDormFloor || floor > state.FloorNum {
		resp := protobuf.SC_19009{Exp: proto.Uint32(0), FoodConsume: proto.Uint32(0)}
		return client.SendMessage(19009, &resp)
	}
	mapSize := dormStaticMapSize(state.Level)
	if err := validateFurniturePutList(request.GetFurniturePutList(), floor, mapSize); err != nil {
		// Client treats this as soft failure; we just avoid persisting and still return zeros.
		resp := protobuf.SC_19009{Exp: proto.Uint32(0), FoodConsume: proto.Uint32(0)}
		return client.SendMessage(19009, &resp)
	}
	stored := make([]storedFurniturePut, 0, len(request.GetFurniturePutList()))
	for _, f := range request.GetFurniturePutList() {
		children := make([]storedChild, 0, len(f.GetChild()))
		for _, c := range f.GetChild() {
			children = append(children, storedChild{Id: c.GetId(), X: c.GetX(), Y: c.GetY()})
		}
		stored = append(stored, storedFurniturePut{
			Id:     f.GetId(),
			X:      f.GetX(),
			Y:      f.GetY(),
			Dir:    f.GetDir(),
			Child:  children,
			Parent: f.GetParent(),
			ShipId: f.GetShipId(),
		})
	}
	b, err := json.Marshal(stored)
	if err != nil {
		return 0, 19009, err
	}
	if err := orm.UpsertCommanderDormFloorLayoutTx(orm.GormDB, commanderID, floor, b); err != nil {
		return 0, 19009, err
	}
	resp := protobuf.SC_19009{Exp: proto.Uint32(0), FoodConsume: proto.Uint32(0)}
	return client.SendMessage(19009, &resp)
}

func ClaimDormIntimacy19011(buffer *[]byte, client *connection.Client) (int, int, error) {
	var request protobuf.CS_19011
	if err := proto.Unmarshal(*buffer, &request); err != nil {
		return 0, 19012, err
	}
	commanderID := client.Commander.CommanderID
	id := request.GetId()

	tx := orm.GormDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if id != 0 {
		var ship orm.OwnedShip
		if err := tx.Where("owner_id = ? AND id = ?", commanderID, id).First(&ship).Error; err != nil {
			tx.Rollback()
			return 0, 19012, err
		}
		if ship.StateInfo3 > 0 {
			ship.Intimacy += ship.StateInfo3
		}
		ship.StateInfo3 = 0
		if err := tx.Save(&ship).Error; err != nil {
			tx.Rollback()
			return 0, 19012, err
		}
	} else {
		var ships []orm.OwnedShip
		if err := tx.Where("owner_id = ? AND state IN (2,5)", commanderID).Find(&ships).Error; err != nil {
			tx.Rollback()
			return 0, 19012, err
		}
		var dormMoney uint32
		for i := range ships {
			dormMoney += ships[i].StateInfo4
			if ships[i].StateInfo3 > 0 {
				ships[i].Intimacy += ships[i].StateInfo3
			}
			ships[i].StateInfo3 = 0
			ships[i].StateInfo4 = 0
			if err := tx.Save(&ships[i]).Error; err != nil {
				tx.Rollback()
				return 0, 19012, err
			}
		}
		if dormMoney > 0 {
			if err := client.Commander.AddResourceTx(tx, 6, dormMoney); err != nil {
				tx.Rollback()
				return 0, 19012, err
			}
		}
	}
	if err := tx.Commit().Error; err != nil {
		return 0, 19012, err
	}
	resp := protobuf.SC_19012{Result: proto.Uint32(0)}
	return client.SendMessage(19012, &resp)
}

func ClaimDormMoney19013(buffer *[]byte, client *connection.Client) (int, int, error) {
	var request protobuf.CS_19013
	if err := proto.Unmarshal(*buffer, &request); err != nil {
		return 0, 19014, err
	}
	commanderID := client.Commander.CommanderID
	shipID := request.GetId()

	tx := orm.GormDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	var ship orm.OwnedShip
	if err := tx.Where("owner_id = ? AND id = ?", commanderID, shipID).First(&ship).Error; err != nil {
		tx.Rollback()
		return 0, 19014, err
	}
	amount := ship.StateInfo4
	ship.StateInfo4 = 0
	if err := tx.Save(&ship).Error; err != nil {
		tx.Rollback()
		return 0, 19014, err
	}
	if amount > 0 {
		if err := client.Commander.AddResourceTx(tx, 6, amount); err != nil {
			tx.Rollback()
			return 0, 19014, err
		}
	}
	if err := tx.Commit().Error; err != nil {
		return 0, 19014, err
	}
	resp := protobuf.SC_19014{Result: proto.Uint32(0)}
	return client.SendMessage(19014, &resp)
}

func OpenAddExp19015(buffer *[]byte, client *connection.Client) (int, int, error) {
	// Client does not expect a direct response.
	// We use this as a tick/poll entrypoint to push dorm pop events.
	_ = buffer
	_ = tickDormAndPush(client)
	return 0, 0, nil
}

func RenameDorm19016(buffer *[]byte, client *connection.Client) (int, int, error) {
	var request protobuf.CS_19016
	if err := proto.Unmarshal(*buffer, &request); err != nil {
		return 0, 19017, err
	}
	client.Commander.DormName = request.GetName()
	if err := orm.GormDB.Save(client.Commander).Error; err != nil {
		resp := protobuf.SC_19017{Result: proto.Uint32(1)}
		return client.SendMessage(19017, &resp)
	}
	resp := protobuf.SC_19017{Result: proto.Uint32(0)}
	return client.SendMessage(19017, &resp)
}

func GetDormThemeList19018(buffer *[]byte, client *connection.Client) (int, int, error) {
	var request protobuf.CS_19018
	if err := proto.Unmarshal(*buffer, &request); err != nil {
		return 0, 19019, err
	}
	commanderID := client.Commander.CommanderID
	requestedID := request.GetId()
	entries, err := orm.ListCommanderDormThemes(commanderID)
	if err != nil {
		return 0, 19019, err
	}
	result := make([]*protobuf.DORMTHEME, 0, len(entries))
	for _, e := range entries {
		if requestedID != 0 && e.ThemeSlotID != requestedID {
			continue
		}
		var stored []storedFurniturePut
		_ = json.Unmarshal(e.FurniturePutList, &stored)
		putList := make([]*protobuf.FURNITUREPUTINFO, 0, len(stored))
		for _, f := range stored {
			children := make([]*protobuf.CHILDINFO, 0, len(f.Child))
			for _, c := range f.Child {
				children = append(children, &protobuf.CHILDINFO{Id: proto.String(c.Id), X: proto.Uint32(c.X), Y: proto.Uint32(c.Y)})
			}
			putList = append(putList, &protobuf.FURNITUREPUTINFO{
				Id:     proto.String(f.Id),
				X:      proto.Uint32(f.X),
				Y:      proto.Uint32(f.Y),
				Dir:    proto.Uint32(f.Dir),
				Child:  children,
				Parent: proto.Uint64(f.Parent),
				ShipId: proto.Uint32(f.ShipId),
			})
		}
		uploadTime := uint32(0)
		result = append(result, &protobuf.DORMTHEME{
			Id:               proto.String(strconv.FormatUint(uint64(e.ThemeSlotID), 10)),
			Name:             proto.String(e.Name),
			UserId:           proto.Uint32(commanderID),
			Pos:              proto.Uint32(e.ThemeSlotID),
			LikeCount:        proto.Uint32(0),
			FavCount:         proto.Uint32(0),
			UploadTime:       proto.Uint32(uploadTime),
			IconImageMd5:     proto.String(""),
			ImageMd5:         proto.String(""),
			FurniturePutList: putList,
		})
	}
	resp := protobuf.SC_19019{ThemeList: result}
	return client.SendMessage(19019, &resp)
}

func SaveDormTheme19020(buffer *[]byte, client *connection.Client) (int, int, error) {
	var request protobuf.CS_19020
	if err := proto.Unmarshal(*buffer, &request); err != nil {
		return 0, 19021, err
	}
	commanderID := client.Commander.CommanderID
	state, err := orm.GetOrCreateCommanderDormState(commanderID)
	if err != nil {
		return 0, 19021, err
	}
	mapSize := dormStaticMapSize(state.Level)
	if err := validateFurniturePutList(request.GetFurniturePutList(), 1, mapSize); err != nil {
		resp := protobuf.SC_19021{Result: proto.Uint32(1)}
		return client.SendMessage(19021, &resp)
	}
	stored := make([]storedFurniturePut, 0, len(request.GetFurniturePutList()))
	for _, f := range request.GetFurniturePutList() {
		children := make([]storedChild, 0, len(f.GetChild()))
		for _, c := range f.GetChild() {
			children = append(children, storedChild{Id: c.GetId(), X: c.GetX(), Y: c.GetY()})
		}
		stored = append(stored, storedFurniturePut{Id: f.GetId(), X: f.GetX(), Y: f.GetY(), Dir: f.GetDir(), Child: children, Parent: f.GetParent(), ShipId: f.GetShipId()})
	}
	b, err := json.Marshal(stored)
	if err != nil {
		return 0, 19021, err
	}
	tx := orm.GormDB.Begin()
	if err := orm.UpsertCommanderDormThemeTx(tx, commanderID, request.GetId(), request.GetName(), b); err != nil {
		tx.Rollback()
		resp := protobuf.SC_19021{Result: proto.Uint32(1)}
		return client.SendMessage(19021, &resp)
	}
	if err := tx.Commit().Error; err != nil {
		resp := protobuf.SC_19021{Result: proto.Uint32(1)}
		return client.SendMessage(19021, &resp)
	}
	resp := protobuf.SC_19021{Result: proto.Uint32(0)}
	return client.SendMessage(19021, &resp)
}

func DeleteDormTheme19022(buffer *[]byte, client *connection.Client) (int, int, error) {
	var request protobuf.CS_19022
	if err := proto.Unmarshal(*buffer, &request); err != nil {
		return 0, 19023, err
	}
	commanderID := client.Commander.CommanderID
	tx := orm.GormDB.Begin()
	if err := orm.DeleteCommanderDormThemeTx(tx, commanderID, request.GetId()); err != nil {
		tx.Rollback()
		resp := protobuf.SC_19023{Result: proto.Uint32(1)}
		return client.SendMessage(19023, &resp)
	}
	if err := tx.Commit().Error; err != nil {
		resp := protobuf.SC_19023{Result: proto.Uint32(1)}
		return client.SendMessage(19023, &resp)
	}
	resp := protobuf.SC_19023{Result: proto.Uint32(0)}
	return client.SendMessage(19023, &resp)
}

func GetBackyardVisitor19024(buffer *[]byte, client *connection.Client) (int, int, error) {
	var request protobuf.CS_19024
	if err := proto.Unmarshal(*buffer, &request); err != nil {
		return 0, 19025, err
	}
	commanderID := client.Commander.CommanderID
	layouts, err := orm.ListCommanderDormFloorLayouts(commanderID)
	if err != nil {
		return 0, 19025, err
	}
	floorPuts := make([]*protobuf.FURFLOORPUTINFO, 0, len(layouts))
	for _, layout := range layouts {
		var stored []storedFurniturePut
		_ = json.Unmarshal(layout.FurniturePutList, &stored)
		putList := make([]*protobuf.FURNITUREPUTINFO, 0, len(stored))
		for _, f := range stored {
			children := make([]*protobuf.CHILDINFO, 0, len(f.Child))
			for _, c := range f.Child {
				children = append(children, &protobuf.CHILDINFO{Id: proto.String(c.Id), X: proto.Uint32(c.X), Y: proto.Uint32(c.Y)})
			}
			putList = append(putList, &protobuf.FURNITUREPUTINFO{Id: proto.String(f.Id), X: proto.Uint32(f.X), Y: proto.Uint32(f.Y), Dir: proto.Uint32(f.Dir), Child: children, Parent: proto.Uint64(f.Parent), ShipId: proto.Uint32(f.ShipId)})
		}
		floorPuts = append(floorPuts, &protobuf.FURFLOORPUTINFO{Floor: proto.Uint32(layout.Floor), FurniturePutList: putList})
	}
	resp := protobuf.SC_19025{FurniturePutList: floorPuts}
	return client.SendMessage(19025, &resp)
}
