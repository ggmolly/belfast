package answer

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func StartLearnTactics(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_22201
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 22202, err
	}

	response := protobuf.SC_22202{Result: proto.Uint32(lessonResultFailed)}
	roomID := payload.GetRoomId()
	shipID := payload.GetShipId()
	skillPos := payload.GetSkillPos()
	itemID := payload.GetItemId()
	if roomID == 0 || shipID == 0 || skillPos == 0 || itemID == 0 {
		return client.SendMessage(22202, &response)
	}

	if client.Commander.OwnedShipsMap == nil || client.Commander.CommanderItemsMap == nil {
		if err := client.Commander.Load(); err != nil {
			return 0, 22202, err
		}
	}

	academyEntries, err := orm.ListConfigEntries("ShareCfg/navalacademy_data_template.json")
	if err != nil {
		return 0, 22202, err
	}
	if roomID > uint32(len(academyEntries)) {
		return client.SendMessage(22202, &response)
	}

	ownedShip, ok := client.Commander.OwnedShipsMap[shipID]
	if !ok {
		return client.SendMessage(22202, &response)
	}

	skillID, err := loadShipSkillByPos(ownedShip.ShipID, skillPos)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return client.SendMessage(22202, &response)
		}
		return 0, 22202, err
	}

	skillConfig, err := loadSkillTemplate(skillID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return client.SendMessage(22202, &response)
		}
		return 0, 22202, err
	}

	itemConfig, usageArg, err := loadLessonItemConfig(itemID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return client.SendMessage(22202, &response)
		}
		return 0, 22202, err
	}
	if itemConfig.Type != lessonItemType || itemConfig.Usage != lessonItemUsage {
		return client.SendMessage(22202, &response)
	}

	duration, lessonExp, err := lessonExpFromUsageArg(usageArg, skillConfig.Type)
	if err != nil {
		return client.SendMessage(22202, &response)
	}

	now := uint32(time.Now().UTC().Unix())
	classInfo := orm.CommanderSkillClass{
		CommanderID: client.Commander.CommanderID,
		RoomID:      roomID,
		ShipID:      shipID,
		SkillPos:    skillPos,
		SkillID:     skillID,
		StartTime:   now,
		FinishTime:  now + duration,
		Exp:         lessonExp,
	}

	ctx := context.Background()
	err = orm.WithPGXTx(ctx, func(tx pgx.Tx) error {
		shipSkill, err := orm.GetOrCreateCommanderShipSkillTx(ctx, tx, client.Commander.CommanderID, shipID, skillPos, skillID)
		if err != nil {
			return err
		}
		if shipSkill.Level >= skillConfig.MaxLevel {
			return db.ErrNotFound
		}
		if err := orm.CreateCommanderSkillClassTx(ctx, tx, &classInfo); err != nil {
			return err
		}
		if err := client.Commander.ConsumeItemTx(ctx, tx, itemID, 1); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, db.ErrNotFound) || errors.Is(err, orm.ErrSkillClassConflict) || err.Error() == "not enough items" {
			return client.SendMessage(22202, &response)
		}
		return 0, 22202, err
	}

	response.Result = proto.Uint32(lessonResultOK)
	response.ClassInfo = &protobuf.SKILL_CLASS{
		RoomId:     proto.Uint32(classInfo.RoomID),
		ShipId:     proto.Uint32(classInfo.ShipID),
		StartTime:  proto.Uint32(classInfo.StartTime),
		FinishTime: proto.Uint32(classInfo.FinishTime),
		SkillPos:   proto.Uint32(classInfo.SkillPos),
		Exp:        proto.Uint32(classInfo.Exp),
	}
	return client.SendMessage(22202, &response)
}
