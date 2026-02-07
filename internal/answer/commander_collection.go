package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"

	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

type MaxShipStat struct {
	GroupID     uint32 `gorm:"column:group_id"`
	MaxStar     uint32 `gorm:"column:max_star"`
	MaxIntimacy uint32 `gorm:"column:max_intimacy"`
	MaxLevel    uint32 `gorm:"column:max_level"`
	MarryFlag   uint32 `gorm:"column:marry_flag"`
	HeartFlag   uint32 `gorm:"column:heart_flag"`
	HeartCount  uint32 `gorm:"column:heart_count"`
}

func CommanderCollection(buffer *[]byte, client *connection.Client) (int, int, error) {
	// Out of all commander's OwnedShips, return the max star, max intimacy, and max level
	// of each ship group (= TemplateID divided by 10)

	var rows []MaxShipStat
	orm.GormDB.Raw(`
	SELECT
		ship_id / 10 AS group_id,
		MAX(ships.star) AS max_star,
		MAX(intimacy) AS max_intimacy,
		MAX(level) AS max_level,
		MAX(CASE WHEN propose THEN 1 ELSE 0 END) AS marry_flag,
		(SELECT COUNT(*) FROM likes WHERE group_id = owned_ships.ship_id / 10 AND liker_id = ?) AS heart_flag,
		(SELECT COUNT(*) FROM likes WHERE group_id = owned_ships.ship_id / 10) AS heart_count
	FROM owned_ships
	INNER JOIN ships ON owned_ships.ship_id = ships.template_id
	WHERE owner_id = ?
	GROUP BY group_id, owned_ships.ship_id
	`, client.Commander.CommanderID, client.Commander.CommanderID).Scan(&rows)

	stats := make([]*protobuf.SHIP_STATISTICS_INFO, len(rows))
	for i, row := range rows {
		stats[i] = &protobuf.SHIP_STATISTICS_INFO{
			Id:          proto.Uint32(row.GroupID),
			Star:        proto.Uint32(row.MaxStar),
			HeartFlag:   proto.Uint32(row.HeartFlag),
			HeartCount:  proto.Uint32(row.HeartCount),
			MarryFlag:   proto.Uint32(row.MarryFlag),
			IntimacyMax: proto.Uint32(row.MaxIntimacy),
			LvMax:       proto.Uint32(row.MaxLevel),
		}
	}

	progress, err := orm.ListCommanderStoreupAwardProgress(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		return 0, 17001, err
	}
	awards := make([]*protobuf.SHIP_STATISTICS_AWARD, 0, len(progress))
	for i := range progress {
		if progress[i].LastAwardIndex == 0 {
			continue
		}
		awards = append(awards, &protobuf.SHIP_STATISTICS_AWARD{
			Id:         proto.Uint32(progress[i].StoreupID),
			AwardIndex: []uint32{progress[i].LastAwardIndex},
		})
	}

	response := protobuf.SC_17001{
		DailyDiscuss:  proto.Uint32(0),
		ShipInfoList:  stats,
		ShipAwardList: awards,
	}
	return client.SendMessage(17001, &response)
}
