package schema

import "gorm.io/gorm"

type Users struct {
	gorm.Model
	Email    string `gorm:"notnull"`
	Password uint   `gorm:"notnull"`
	Username string
	Avatar   []byte
}
