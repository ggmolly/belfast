package orm

type Server struct {
	ID        uint32  `gorm:"primary_key"`
	IP        string  `gorm:"size:255;not_null"`
	Port      uint32  `gorm:"not_null"`
	Name      string  `gorm:"size:64;not_null"`
	StateID   *uint32 `gorm:"not_null"`
	ProxyIP   *string `gorm:"size:255"`
	ProxyPort *int

	State ServerState `gorm:"foreignKey:StateID"`
}

type ServerState struct {
	ID          uint32   `gorm:"primary_key"`
	Description string   `gorm:"size:10;default:'Unknown';not_null"`
	Color       string   `gorm:"size:8;not_null"`
	Servers     []Server `gorm:"foreignKey:StateID"`
}

// Inserts or updates a server in the database (based on the primary key)
func (s *Server) Create() error {
	return GormDB.Save(s).Error
}

// Updates a server in the database
func (s *Server) Update() error {
	return GormDB.Model(s).Updates(s).Error
}

// Gets a server from the database by its primary key
// If greedy is true, it will also load the relations
func (s *Server) Retrieve(greedy bool) error {
	if greedy {
		return GormDB.
			Joins("JOIN server_states ON server_states.id = servers.state_id").
			Where("servers.id = ?", s.ID).
			First(s).Error
	} else {
		return GormDB.
			Where("id = ?", s.ID).
			First(s).Error
	}
}

// Deletes a server from the database
func (s *Server) Delete() error {
	return GormDB.Delete(s).Error
}
