package orm

import (
	"fmt"

	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

// OwnedSpWeapon represents a special weapon (spweapon) owned by a commander.
//
// The client refers to spweapons by a per-instance uid; we persist that uid as
// the primary key ID.
type OwnedSpWeapon struct {
	OwnerID uint32 `gorm:"type:int;not_null;index"`
	ID      uint32 `gorm:"primary_key"`

	TemplateID uint32 `gorm:"not_null"`

	Attr1     uint32 `gorm:"column:attr_1;default:0;not_null"`
	Attr2     uint32 `gorm:"column:attr_2;default:0;not_null"`
	AttrTemp1 uint32 `gorm:"column:attr_temp_1;default:0;not_null"`
	AttrTemp2 uint32 `gorm:"column:attr_temp_2;default:0;not_null"`
	Effect    uint32 `gorm:"default:0;not_null"`
	Pt        uint32 `gorm:"default:0;not_null"`

	// EquippedShipID is the owned ship id this spweapon is equipped to, or 0.
	EquippedShipID uint32 `gorm:"default:0;not_null"`
}

func (c *Commander) ensureOwnedSpWeaponMap() {
	if c.OwnedSpWeaponsMap == nil {
		c.rebuildOwnedSpWeaponMap()
	}
}

func (c *Commander) rebuildOwnedSpWeaponMap() {
	c.OwnedSpWeaponsMap = make(map[uint32]*OwnedSpWeapon, len(c.OwnedSpWeapons))
	for i := range c.OwnedSpWeapons {
		c.OwnedSpWeaponsMap[c.OwnedSpWeapons[i].ID] = &c.OwnedSpWeapons[i]
	}
}

func (c *Commander) RebuildOwnedSpWeaponMap() {
	c.rebuildOwnedSpWeaponMap()
}

func (OwnedSpWeapon) TableName() string {
	return "owned_spweapons"
}

func ListOwnedSpWeapons(db *gorm.DB, ownerID uint32) ([]OwnedSpWeapon, error) {
	var entries []OwnedSpWeapon
	if err := db.Where("owner_id = ?", ownerID).Order("id asc").Find(&entries).Error; err != nil {
		return nil, err
	}
	return entries, nil
}

func GetOwnedSpWeapon(db *gorm.DB, ownerID uint32, spweaponID uint32) (*OwnedSpWeapon, error) {
	var entry OwnedSpWeapon
	if err := db.Where("owner_id = ? AND id = ?", ownerID, spweaponID).First(&entry).Error; err != nil {
		return nil, err
	}
	return &entry, nil
}

func UpsertOwnedSpWeaponTx(tx *gorm.DB, entry *OwnedSpWeapon) error {
	return tx.Save(entry).Error
}

func (c *Commander) RemoveOwnedSpWeaponTx(tx *gorm.DB, spweaponID uint32) error {
	c.ensureOwnedSpWeaponMap()
	if _, ok := c.OwnedSpWeaponsMap[spweaponID]; !ok {
		return fmt.Errorf("spweapon not owned")
	}
	if err := tx.Where("owner_id = ? AND id = ?", c.CommanderID, spweaponID).Delete(&OwnedSpWeapon{}).Error; err != nil {
		return err
	}
	for i := range c.OwnedSpWeapons {
		if c.OwnedSpWeapons[i].ID == spweaponID {
			c.OwnedSpWeapons = append(c.OwnedSpWeapons[:i], c.OwnedSpWeapons[i+1:]...)
			break
		}
	}
	c.rebuildOwnedSpWeaponMap()
	return nil
}

func ToProtoOwnedSpWeapon(entry OwnedSpWeapon) *protobuf.SPWEAPONINFO {
	return &protobuf.SPWEAPONINFO{
		Id:         proto.Uint32(entry.ID),
		TemplateId: proto.Uint32(entry.TemplateID),
		Attr_1:     proto.Uint32(entry.Attr1),
		Attr_2:     proto.Uint32(entry.Attr2),
		AttrTemp_1: proto.Uint32(entry.AttrTemp1),
		AttrTemp_2: proto.Uint32(entry.AttrTemp2),
		Effect:     proto.Uint32(entry.Effect),
		Pt:         proto.Uint32(entry.Pt),
	}
}

func ToProtoOwnedSpWeaponList(entries []OwnedSpWeapon) []*protobuf.SPWEAPONINFO {
	if len(entries) == 0 {
		return nil
	}
	result := make([]*protobuf.SPWEAPONINFO, len(entries))
	for i, entry := range entries {
		result[i] = ToProtoOwnedSpWeapon(entry)
	}
	return result
}
