package answer

import (
	"os"
	"strconv"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm/clause"
)

const (
	equipCodeImpeachResultOK  uint32 = 0
	equipCodeImpeachResultErr uint32 = 1

	equipCodeImpeachDefaultDailyLimit uint32 = 5
)

const equipCodeImpeachWarningResult uint32 = ^uint32(0)

func equipCodeImpeachDailyLimit() uint32 {
	raw := os.Getenv("EQUIP_CODE_IMPEACH_DAILY_LIMIT")
	if raw == "" {
		return equipCodeImpeachDefaultDailyLimit
	}
	limit, err := strconv.ParseUint(raw, 10, 32)
	if err != nil || limit == 0 {
		return equipCodeImpeachDefaultDailyLimit
	}
	return uint32(limit)
}

func EquipCodeImpeach(buffer *[]byte, client *connection.Client) (int, int, error) {
	var data protobuf.CS_17607
	if err := proto.Unmarshal(*buffer, &data); err != nil {
		return 0, 17608, err
	}

	response := protobuf.SC_17608{Result: proto.Uint32(equipCodeImpeachResultOK)}
	shipGroupID := data.GetShipgroup()
	shareID := data.GetShareid()
	reportType := data.GetReportType()
	if shipGroupID == 0 || shareID == 0 || (reportType != 1 && reportType != 2) {
		response.Result = proto.Uint32(equipCodeImpeachResultErr)
		return client.SendMessage(17608, &response)
	}

	now := time.Now().UTC()
	day := uint32(now.Unix() / 86400)
	commanderID := client.Commander.CommanderID

	report := orm.EquipCodeReport{
		CommanderID: commanderID,
		ShipGroupID: shipGroupID,
		ShareID:     shareID,
		ReportType:  reportType,
		ReportDay:   day,
		CreatedAt:   now,
	}
	if err := orm.GormDB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "commander_id"}, {Name: "share_id"}, {Name: "report_day"}},
		DoNothing: true,
	}).Create(&report).Error; err != nil {
		return 0, 17608, err
	}

	limit := equipCodeImpeachDailyLimit()
	since := now.Add(-24 * time.Hour)
	var count int64
	if err := orm.GormDB.Model(&orm.EquipCodeReport{}).
		Where("commander_id = ? AND created_at >= ?", commanderID, since).
		Count(&count).Error; err != nil {
		return 0, 17608, err
	}
	if uint32(count) > limit {
		response.Result = proto.Uint32(equipCodeImpeachWarningResult)
	}

	return client.SendMessage(17608, &response)
}
