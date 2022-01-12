package schema

import "gorm.io/gorm"

type Team struct {
	gorm.Model
	FirstID    uint
	SecondID   uint
	FirstUser  User `gorm:"foreignKey:FirstID"`
	SecondUser User `gorm:"foreignKey:SecondID"`
}
