package schema

import "gorm.io/gorm"

type PlayerWord struct {
	gorm.Model
	UserID uint
	WordID uint
}
