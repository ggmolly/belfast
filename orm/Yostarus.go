package orm

// Implementation of dumb auth
// When logging-in, the client sends a packet with the field arg2, which seems to
// be unique and tied to an account id.
type YostarusMap struct {
	Arg2      uint32 `gorm:"primary_key"`
	AccountID uint32 `gorm:"not_null;uniqueIndex"`

	Commander Commander `gorm:"foreignkey:AccountID;references:AccountID"`
}
