package schema

import "gorm.io/gorm"

type PlayerWord struct {
	gorm.Model
	AuthorID    uint
	GuessedByID uint
	Author      User `gorm:"foreignKey:AuthorID"`
	GuessedBy   User `gorm:"foreignKey:GuessedByID"`
	WordID      uint
}
