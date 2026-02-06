package orm

import (
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// OwnedShipSkinShadow persists which phantom/shadow slots are unlocked for an owned ship.
// This maps to SHIPINFO.skin_shadow_list entries (KVDATA{key=shadow_id, value=skin_id}).
type OwnedShipSkinShadow struct {
	CommanderID uint32 `gorm:"primaryKey;autoIncrement:false"`
	ShipID      uint32 `gorm:"primaryKey;autoIncrement:false"`
	ShadowID    uint32 `gorm:"primaryKey;autoIncrement:false"`
	SkinID      uint32 `gorm:"not_null;default:0"`
}

func UpsertOwnedShipSkinShadow(db *gorm.DB, entry *OwnedShipSkinShadow) error {
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "commander_id"}, {Name: "ship_id"}, {Name: "shadow_id"}},
		DoNothing: true,
	}).Create(entry).Error
}

func ListOwnedShipSkinShadows(commanderID uint32, shipIDs []uint32) (map[uint32][]*protobuf.KVDATA, error) {
	result := make(map[uint32][]*protobuf.KVDATA)
	var entries []OwnedShipSkinShadow
	query := GormDB.Where("commander_id = ?", commanderID)
	if len(shipIDs) > 0 {
		query = query.Where("ship_id IN ?", shipIDs)
	}
	if err := query.Order("ship_id asc").Order("shadow_id asc").Find(&entries).Error; err != nil {
		return nil, err
	}
	for _, entry := range entries {
		result[entry.ShipID] = append(result[entry.ShipID], &protobuf.KVDATA{
			Key:   proto.Uint32(entry.ShadowID),
			Value: proto.Uint32(entry.SkinID),
		})
	}
	return result, nil
}
