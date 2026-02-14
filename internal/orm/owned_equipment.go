package orm

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type OwnedEquipment struct {
	CommanderID uint32 `gorm:"primaryKey;autoIncrement:false" json:"commander_id"`
	EquipmentID uint32 `gorm:"primaryKey;autoIncrement:false" json:"equipment_id"`
	Count       uint32 `gorm:"not_null" json:"count"`

	Equipment Equipment `gorm:"foreignKey:EquipmentID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

func (c *Commander) EquipmentBagCount() uint32 {
	var total uint32
	for _, entry := range c.OwnedEquipments {
		total += entry.Count
	}
	return total
}

func (c *Commander) GetOwnedEquipment(equipmentID uint32) *OwnedEquipment {
	c.ensureOwnedEquipmentMap()
	return c.OwnedEquipmentMap[equipmentID]
}

func (c *Commander) AddOwnedEquipmentTx(ctx context.Context, tx pgx.Tx, equipmentID uint32, count uint32) error {
	if count == 0 {
		return nil
	}
	c.ensureOwnedEquipmentMap()
	if existing, ok := c.OwnedEquipmentMap[equipmentID]; ok {
		_, err := tx.Exec(ctx, `
UPDATE owned_equipments
SET count = count + $3
WHERE commander_id = $1 AND equipment_id = $2
`, int64(c.CommanderID), int64(equipmentID), int64(count))
		if err != nil {
			return err
		}
		existing.Count += count
		return nil
	}
	entry := OwnedEquipment{CommanderID: c.CommanderID, EquipmentID: equipmentID, Count: count}
	if _, err := tx.Exec(ctx, `
INSERT INTO owned_equipments (commander_id, equipment_id, count)
VALUES ($1, $2, $3)
`, int64(entry.CommanderID), int64(entry.EquipmentID), int64(entry.Count)); err != nil {
		return err
	}
	c.OwnedEquipments = append(c.OwnedEquipments, entry)
	c.rebuildOwnedEquipmentMap()
	return nil
}

func (c *Commander) RemoveOwnedEquipmentTx(ctx context.Context, tx pgx.Tx, equipmentID uint32, count uint32) error {
	if count == 0 {
		return nil
	}
	c.ensureOwnedEquipmentMap()
	existing, ok := c.OwnedEquipmentMap[equipmentID]
	if !ok || existing.Count < count {
		return fmt.Errorf("not enough equipment")
	}
	existing.Count -= count
	if existing.Count == 0 {
		if _, err := tx.Exec(ctx, `
DELETE FROM owned_equipments
WHERE commander_id = $1 AND equipment_id = $2
`, int64(c.CommanderID), int64(equipmentID)); err != nil {
			return err
		}
		for i := range c.OwnedEquipments {
			if c.OwnedEquipments[i].EquipmentID == equipmentID {
				c.OwnedEquipments = append(c.OwnedEquipments[:i], c.OwnedEquipments[i+1:]...)
				break
			}
		}
		c.rebuildOwnedEquipmentMap()
		return nil
	}
	_, err := tx.Exec(ctx, `
UPDATE owned_equipments
SET count = $3
WHERE commander_id = $1 AND equipment_id = $2
`, int64(c.CommanderID), int64(equipmentID), int64(existing.Count))
	return err
}

func (c *Commander) SetOwnedEquipmentTx(ctx context.Context, tx pgx.Tx, equipmentID uint32, count uint32) error {
	if count == 0 {
		if _, err := tx.Exec(ctx, `
DELETE FROM owned_equipments
WHERE commander_id = $1 AND equipment_id = $2
`, int64(c.CommanderID), int64(equipmentID)); err != nil {
			return err
		}
		for i := range c.OwnedEquipments {
			if c.OwnedEquipments[i].EquipmentID == equipmentID {
				c.OwnedEquipments = append(c.OwnedEquipments[:i], c.OwnedEquipments[i+1:]...)
				break
			}
		}
		c.rebuildOwnedEquipmentMap()
		return nil
	}
	c.ensureOwnedEquipmentMap()
	if existing, ok := c.OwnedEquipmentMap[equipmentID]; ok {
		_, err := tx.Exec(ctx, `
UPDATE owned_equipments
SET count = $3
WHERE commander_id = $1 AND equipment_id = $2
`, int64(c.CommanderID), int64(equipmentID), int64(count))
		if err != nil {
			return err
		}
		existing.Count = count
		return nil
	}
	entry := OwnedEquipment{CommanderID: c.CommanderID, EquipmentID: equipmentID, Count: count}
	if _, err := tx.Exec(ctx, `
INSERT INTO owned_equipments (commander_id, equipment_id, count)
VALUES ($1, $2, $3)
`, int64(entry.CommanderID), int64(entry.EquipmentID), int64(entry.Count)); err != nil {
		return err
	}
	c.OwnedEquipments = append(c.OwnedEquipments, entry)
	c.rebuildOwnedEquipmentMap()
	return nil
}

func (c *Commander) ensureOwnedEquipmentMap() {
	if c.OwnedEquipmentMap == nil {
		c.rebuildOwnedEquipmentMap()
	}
}

func (c *Commander) rebuildOwnedEquipmentMap() {
	c.OwnedEquipmentMap = make(map[uint32]*OwnedEquipment, len(c.OwnedEquipments))
	for i := range c.OwnedEquipments {
		c.OwnedEquipmentMap[c.OwnedEquipments[i].EquipmentID] = &c.OwnedEquipments[i]
	}
}

func (c *Commander) RebuildOwnedEquipmentMap() {
	c.rebuildOwnedEquipmentMap()
}
