package answer

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm/clause"
)

const (
	equipCodeShareResultOK uint32 = 0
	// Generic failure for invalid payload/validation errors.
	equipCodeShareResultErr uint32 = 1

	// Client semantics:
	// 7  -> already shared this ship group's loadout today
	// 44 -> global daily share limit reached
	equipCodeShareResultAlreadyShared uint32 = 7
	equipCodeShareResultDailyLimit    uint32 = 44

	equipCodeShareDefaultDailyLimit uint32 = 5
)

func equipCodeShareDailyLimit() uint32 {
	raw := os.Getenv("EQUIP_CODE_SHARE_DAILY_LIMIT")
	if raw == "" {
		return equipCodeShareDefaultDailyLimit
	}
	limit, err := strconv.ParseUint(raw, 10, 32)
	if err != nil || limit == 0 {
		return equipCodeShareDefaultDailyLimit
	}
	return uint32(limit)
}

func decodeConversionBase32(s string) (uint32, bool) {
	if s == "" {
		return 0, false
	}

	var out uint32
	for i := 0; i < len(s); i++ {
		c := s[i]
		var digit uint32
		switch {
		case c >= '0' && c <= '9':
			digit = uint32(c - '0')
		case c >= 'a' && c <= 'z':
			c = c - ('a' - 'A')
			fallthrough
		case c >= 'A' && c <= 'Z':
			digit = uint32(c-'A') + 10
		default:
			return 0, false
		}
		if digit >= 32 {
			return 0, false
		}
		out = out*32 + digit
	}
	return out, true
}

func validateEquipSharePayload(shipGroupID uint32, eqcode string) bool {
	if shipGroupID == 0 || eqcode == "" {
		return false
	}
	parts := strings.Split(eqcode, "&")
	if len(parts) != 4 {
		return false
	}
	encodedGroup := parts[1]
	decoded, ok := decodeConversionBase32(encodedGroup)
	if !ok {
		return false
	}
	return decoded == shipGroupID
}

func EquipCodeShare(buffer *[]byte, client *connection.Client) (int, int, error) {
	var data protobuf.CS_17603
	if err := proto.Unmarshal(*buffer, &data); err != nil {
		return 0, 17604, err
	}

	response := protobuf.SC_17604{Result: proto.Uint32(equipCodeShareResultOK)}
	shipGroupID := data.GetShipgroup()
	eqcode := data.GetEqcode()
	if !validateEquipSharePayload(shipGroupID, eqcode) {
		response.Result = proto.Uint32(equipCodeShareResultErr)
		return client.SendMessage(17604, &response)
	}

	now := time.Now().UTC()
	day := uint32(now.Unix() / 86400)
	commanderID := client.Commander.CommanderID

	var alreadyShared int64
	if err := orm.GormDB.Model(&orm.EquipCodeShare{}).
		Where("commander_id = ? AND ship_group_id = ? AND share_day = ?", commanderID, shipGroupID, day).
		Count(&alreadyShared).Error; err != nil {
		return 0, 17604, err
	}
	if alreadyShared > 0 {
		response.Result = proto.Uint32(equipCodeShareResultAlreadyShared)
		return client.SendMessage(17604, &response)
	}

	limit := equipCodeShareDailyLimit()
	var totalShares int64
	if err := orm.GormDB.Model(&orm.EquipCodeShare{}).
		Where("commander_id = ? AND share_day = ?", commanderID, day).
		Count(&totalShares).Error; err != nil {
		return 0, 17604, err
	}
	if uint32(totalShares) >= limit {
		response.Result = proto.Uint32(equipCodeShareResultDailyLimit)
		return client.SendMessage(17604, &response)
	}

	share := orm.EquipCodeShare{
		CommanderID: commanderID,
		ShipGroupID: shipGroupID,
		ShareDay:    day,
		CreatedAt:   now,
	}
	// Dedupe on (commander_id, ship_group_id, share_day) to ensure the client sees
	// the correct per-shipgroup result even under concurrent requests.
	tx := orm.GormDB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "commander_id"}, {Name: "ship_group_id"}, {Name: "share_day"}},
		DoNothing: true,
	}).Create(&share)
	if tx.Error != nil {
		return 0, 17604, tx.Error
	}
	if tx.RowsAffected == 0 {
		response.Result = proto.Uint32(equipCodeShareResultAlreadyShared)
	}

	return client.SendMessage(17604, &response)
}
