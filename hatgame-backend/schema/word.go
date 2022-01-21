package schema

import "gorm.io/gorm"

type Word struct {
	gorm.Model
	Word string `gorm:"unique,notnull"`
}
