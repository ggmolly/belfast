package orm

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
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

func ListOwnedSpWeapons(ownerID uint32) ([]OwnedSpWeapon, error) {
	ctx := context.Background()
	rows, err := db.DefaultStore.Pool.Query(ctx, `
SELECT owner_id, id, template_id, attr_1, attr_2, attr_temp_1, attr_temp_2, effect, pt, equipped_ship_id
FROM owned_spweapons
WHERE owner_id = $1
ORDER BY id ASC
`, int64(ownerID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := make([]OwnedSpWeapon, 0)
	for rows.Next() {
		var entry OwnedSpWeapon
		if err := rows.Scan(&entry.OwnerID, &entry.ID, &entry.TemplateID, &entry.Attr1, &entry.Attr2, &entry.AttrTemp1, &entry.AttrTemp2, &entry.Effect, &entry.Pt, &entry.EquippedShipID); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

func GetOwnedSpWeapon(ownerID uint32, spweaponID uint32) (*OwnedSpWeapon, error) {
	ctx := context.Background()
	row := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT owner_id, id, template_id, attr_1, attr_2, attr_temp_1, attr_temp_2, effect, pt, equipped_ship_id
FROM owned_spweapons
WHERE owner_id = $1 AND id = $2
`, int64(ownerID), int64(spweaponID))

	var entry OwnedSpWeapon
	err := row.Scan(&entry.OwnerID, &entry.ID, &entry.TemplateID, &entry.Attr1, &entry.Attr2, &entry.AttrTemp1, &entry.AttrTemp2, &entry.Effect, &entry.Pt, &entry.EquippedShipID)
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func UpsertOwnedSpWeaponTx(ctx context.Context, tx pgx.Tx, entry *OwnedSpWeapon) error {
	_, err := tx.Exec(ctx, `
INSERT INTO owned_spweapons (
  owner_id,
  id,
  template_id,
  attr_1,
  attr_2,
  attr_temp_1,
  attr_temp_2,
  effect,
  pt,
  equipped_ship_id
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
ON CONFLICT (id)
DO UPDATE SET
  template_id = EXCLUDED.template_id,
  attr_1 = EXCLUDED.attr_1,
  attr_2 = EXCLUDED.attr_2,
  attr_temp_1 = EXCLUDED.attr_temp_1,
  attr_temp_2 = EXCLUDED.attr_temp_2,
  effect = EXCLUDED.effect,
  pt = EXCLUDED.pt,
  equipped_ship_id = EXCLUDED.equipped_ship_id
WHERE owned_spweapons.owner_id = EXCLUDED.owner_id
`, int64(entry.OwnerID), int64(entry.ID), int64(entry.TemplateID), int64(entry.Attr1), int64(entry.Attr2), int64(entry.AttrTemp1), int64(entry.AttrTemp2), int64(entry.Effect), int64(entry.Pt), int64(entry.EquippedShipID))
	return err
}

func (c *Commander) RemoveOwnedSpWeaponTx(ctx context.Context, tx pgx.Tx, spweaponID uint32) error {
	c.ensureOwnedSpWeaponMap()
	if _, ok := c.OwnedSpWeaponsMap[spweaponID]; !ok {
		return fmt.Errorf("spweapon not owned")
	}
	if _, err := tx.Exec(ctx, `DELETE FROM owned_spweapons WHERE owner_id = $1 AND id = $2`, int64(c.CommanderID), int64(spweaponID)); err != nil {
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

func SaveOwnedSpWeapon(entry *OwnedSpWeapon) error {
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `
UPDATE owned_spweapons
SET
  template_id = $3,
  attr_1 = $4,
  attr_2 = $5,
  attr_temp_1 = $6,
  attr_temp_2 = $7,
  effect = $8,
  pt = $9,
  equipped_ship_id = $10
WHERE owner_id = $1 AND id = $2
`, int64(entry.OwnerID), int64(entry.ID), int64(entry.TemplateID), int64(entry.Attr1), int64(entry.Attr2), int64(entry.AttrTemp1), int64(entry.AttrTemp2), int64(entry.Effect), int64(entry.Pt), int64(entry.EquippedShipID))
	return err
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

func CreateOwnedSpWeapon(ownerID uint32, templateID uint32) (*OwnedSpWeapon, error) {
	ctx := context.Background()
	row := db.DefaultStore.Pool.QueryRow(ctx, `
INSERT INTO owned_spweapons (
  owner_id,
  template_id,
  attr_1,
  attr_2,
  attr_temp_1,
  attr_temp_2,
  effect,
  pt,
  equipped_ship_id
) VALUES (
  $1, $2, 0, 0, 0, 0, 0, 0, 0
)
RETURNING owner_id, id, template_id, attr_1, attr_2, attr_temp_1, attr_temp_2, effect, pt, equipped_ship_id
`, int64(ownerID), int64(templateID))
	var entry OwnedSpWeapon
	if err := row.Scan(&entry.OwnerID, &entry.ID, &entry.TemplateID, &entry.Attr1, &entry.Attr2, &entry.AttrTemp1, &entry.AttrTemp2, &entry.Effect, &entry.Pt, &entry.EquippedShipID); err != nil {
		return nil, err
	}
	return &entry, nil
}
