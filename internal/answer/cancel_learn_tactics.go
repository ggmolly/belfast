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

func CancelLearnTactics(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_22203
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 22204, err
	}

	response := protobuf.SC_22204{Result: proto.Uint32(lessonResultFailed)}
	roomID := payload.GetRoomId()
	cancelType := payload.GetType()
	if roomID == 0 {
		return client.SendMessage(22204, &response)
	}
	if cancelType != skillCancelTypeAuto && cancelType != skillCancelTypeManual {
		return client.SendMessage(22204, &response)
	}

	var grantedExp uint32
	ctx := context.Background()
	err := orm.WithPGXTx(ctx, func(tx pgx.Tx) error {
		lesson, err := orm.GetCommanderSkillClassByRoomTx(ctx, tx, client.Commander.CommanderID, roomID)
		if err != nil {
			return err
		}
		skillCfg, err := loadSkillTemplate(lesson.SkillID)
		if err != nil {
			return err
		}
		shipSkill, err := orm.GetOrCreateCommanderShipSkillTx(ctx, tx, client.Commander.CommanderID, lesson.ShipID, lesson.SkillPos, lesson.SkillID)
		if err != nil {
			return err
		}
		candidateExp := calcGrantedLessonExp(time.Now().UTC(), lesson.StartTime, lesson.FinishTime, lesson.Exp)
		grantedExp = applyLessonExp(shipSkill, candidateExp, skillCfg.MaxLevel)
		if err := orm.SaveCommanderShipSkillTx(ctx, tx, shipSkill); err != nil {
			return err
		}
		return orm.DeleteCommanderSkillClassTx(ctx, tx, client.Commander.CommanderID, roomID)
	})
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return client.SendMessage(22204, &response)
		}
		return 0, 22204, err
	}

	response.Result = proto.Uint32(lessonResultOK)
	response.Exp = proto.Uint32(grantedExp)
	return client.SendMessage(22204, &response)
}
