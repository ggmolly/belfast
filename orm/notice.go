package orm

type Notice struct {
	ID            int    `gorm:"primary_key"`
	Version       int    `gorm:"type:smallint;default:1;not_null"`
	BtnTitle      string `gorm:"size:128;not_null"`
	Title         string `gorm:"size:256;not_null"`
	TitleImageURL string `gorm:"type:text;not_null"`
	TimeDesc      string `gorm:"size:10;not_null"`
	Content       string `gorm:"type:text;not_null"`
}

// Inserts or updates a notice in the database (based on the primary key)
func (n *Notice) Create() error {
	return GormDB.Save(n).Error
}

// Updates a notice in the database
func (n *Notice) Update() error {
	return GormDB.Model(n).Updates(n).Error
}

// Gets a notice from the database by its primary key
// If greedy is true, it will also load the relations
func (n *Notice) Retrieve(greedy bool) error {
	// ignore greediness because there are no relations
	return GormDB.
		Where("id = ?", n.ID).
		First(n).Error
}

// Deletes a notice from the database
func (n *Notice) Delete() error {
	return GormDB.Delete(n).Error
}
