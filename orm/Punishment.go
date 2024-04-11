package orm

import "time"

type Punishment struct {
	ID            uint32     `gorm:"primary_key"`
	PunishedID    uint32     `gorm:"not_null"`
	LiftTimestamp *time.Time `gorm:"type:timestamp"`
	IsPermanent   bool       `gorm:"default:false;not_null"`

	Punished Commander `gorm:"foreignKey:PunishedID;references:CommanderID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

// Inserts or updates a punishment in the database (based on the primary key)
func (p *Punishment) Create() error {
	return GormDB.Save(p).Error
}

// Updates a punishment in the database
func (p *Punishment) Update() error {
	return GormDB.Model(p).Updates(p).Error
}

// Gets a punishment from the database by its primary key
// If greedy is true, it will also load the relations
func (p *Punishment) Retrieve(greedy bool) error {
	if greedy {
		return GormDB.
			Joins("JOIN commanders ON commanders.commander_id = punishments.punished_id").
			Where("punishments.id = ?", p.ID).
			First(p).Error
	} else {
		return GormDB.
			Where("id = ?", p.ID).
			First(p).Error
	}
}

// Deletes a punishment from the database
func (p *Punishment) Delete() error {
	return GormDB.Delete(p).Error
}
