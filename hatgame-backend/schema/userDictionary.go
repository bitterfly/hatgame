package schema

import "gorm.io/gorm"

type UserDictionary struct {
	gorm.Model
	AuthorID uint
	WordID   uint
	Author   User `gorm:"foreignKey:AuthorID"`
	Word     Word
}
