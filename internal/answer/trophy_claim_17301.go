package answer

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

const medalTemplateCategory = "ShareCfg/medal_template.json"

var errTrophyClaimRejected = errors.New("trophy claim rejected")

type medalTemplate struct {
	ID           uint32 `json:"id"`
	Next         uint32 `json:"next"`
	TargetNum    uint32 `json:"target_num"`
	CountInherit uint32 `json:"count_inherit"`
}

func TrophyClaim17301(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_17301
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 17302, err
	}

	response := protobuf.SC_17302{Result: proto.Uint32(1)}
	medalID := payload.GetId()
	if medalID == 0 {
		return client.SendMessage(17302, &response)
	}

	entry, err := orm.GetConfigEntry(orm.GormDB, medalTemplateCategory, strconv.FormatUint(uint64(medalID), 10))
	if err != nil {
		return client.SendMessage(17302, &response)
	}
	var template medalTemplate
	if err := json.Unmarshal(entry.Data, &template); err != nil {
		return 0, 17302, err
	}

	commanderID := client.Commander.CommanderID
	var claimTimestamp uint32
	var unlockedNext *protobuf.ACHIEVEMENT_INFO
	if err := orm.GormDB.Transaction(func(tx *gorm.DB) error {
		trophy, _, err := orm.GetOrCreateCommanderTrophyProgress(tx, commanderID, medalID, template.TargetNum)
		if err != nil {
			return err
		}
		if trophy.Timestamp != 0 {
			return errTrophyClaimRejected
		}
		if trophy.Progress < template.TargetNum {
			return errTrophyClaimRejected
		}

		claimTimestamp = uint32(time.Now().Unix())
		if err := orm.ClaimCommanderTrophyProgress(tx, commanderID, medalID, claimTimestamp); err != nil {
			return err
		}

		nextID := template.Next
		if nextID == 0 {
			return nil
		}
		nextProgress := uint32(0)
		if template.CountInherit == nextID {
			nextProgress = trophy.Progress
		}
		nextRow, created, err := orm.GetOrCreateCommanderTrophyProgress(tx, commanderID, nextID, nextProgress)
		if err != nil {
			return err
		}
		if created {
			unlockedNext = &protobuf.ACHIEVEMENT_INFO{
				Id:        proto.Uint32(nextRow.TrophyID),
				Progress:  proto.Uint32(nextRow.Progress),
				Timestamp: proto.Uint32(nextRow.Timestamp),
			}
		}
		return nil
	}); err != nil {
		if errors.Is(err, errTrophyClaimRejected) {
			return client.SendMessage(17302, &response)
		}
		return 0, 17302, err
	}

	response.Result = proto.Uint32(0)
	response.Timestamp = proto.Uint32(claimTimestamp)
	if unlockedNext != nil {
		response.Next = []*protobuf.ACHIEVEMENT_INFO{unlockedNext}
	}
	return client.SendMessage(17302, &response)
}
