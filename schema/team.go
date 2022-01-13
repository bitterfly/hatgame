package schema

import "gorm.io/gorm"

type Team struct {
	gorm.Model
	FirstID    uint `gorm:"index:idx_name,unique"`
	SecondID   uint `gorm:"index:idx_name,unique"`
	FirstUser  User `gorm:"foreignKey:FirstID"`
	SecondUser User `gorm:"foreignKey:SecondID"`
}
