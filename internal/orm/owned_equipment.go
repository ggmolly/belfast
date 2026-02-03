package orm

import (
	"fmt"

	"gorm.io/gorm"
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

func (c *Commander) AddOwnedEquipmentTx(tx *gorm.DB, equipmentID uint32, count uint32) error {
	if count == 0 {
		return nil
	}
	c.ensureOwnedEquipmentMap()
	if existing, ok := c.OwnedEquipmentMap[equipmentID]; ok {
		existing.Count += count
		return tx.Save(existing).Error
	}
	entry := OwnedEquipment{CommanderID: c.CommanderID, EquipmentID: equipmentID, Count: count}
	if err := tx.Create(&entry).Error; err != nil {
		return err
	}
	c.OwnedEquipments = append(c.OwnedEquipments, entry)
	c.rebuildOwnedEquipmentMap()
	return nil
}

func (c *Commander) RemoveOwnedEquipmentTx(tx *gorm.DB, equipmentID uint32, count uint32) error {
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
		if err := tx.Where("commander_id = ? AND equipment_id = ?", c.CommanderID, equipmentID).Delete(&OwnedEquipment{}).Error; err != nil {
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
	return tx.Save(existing).Error
}

func (c *Commander) SetOwnedEquipmentTx(tx *gorm.DB, equipmentID uint32, count uint32) error {
	if count == 0 {
		if err := tx.Where("commander_id = ? AND equipment_id = ?", c.CommanderID, equipmentID).Delete(&OwnedEquipment{}).Error; err != nil {
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
		existing.Count = count
		return tx.Save(existing).Error
	}
	entry := OwnedEquipment{CommanderID: c.CommanderID, EquipmentID: equipmentID, Count: count}
	if err := tx.Create(&entry).Error; err != nil {
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
