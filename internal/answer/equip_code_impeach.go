package answer

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
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
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `
INSERT INTO equip_code_reports (commander_id, ship_group_id, share_id, report_type, report_day, created_at)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (commander_id, share_id, report_day)
DO NOTHING
`, int64(report.CommanderID), int64(report.ShipGroupID), int64(report.ShareID), int64(report.ReportType), int64(report.ReportDay), report.CreatedAt); err != nil {
		return 0, 17608, err
	}

	limit := equipCodeImpeachDailyLimit()
	since := now.Add(-24 * time.Hour)
	var count int64
	if err := db.DefaultStore.Pool.QueryRow(context.Background(), `
SELECT COUNT(*)
FROM equip_code_reports
WHERE commander_id = $1
  AND created_at >= $2
`, int64(commanderID), since).Scan(&count); err != nil {
		return 0, 17608, err
	}
	if uint32(count) > limit {
		response.Result = proto.Uint32(equipCodeImpeachWarningResult)
	}

	return client.SendMessage(17608, &response)
}
