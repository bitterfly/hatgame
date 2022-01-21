package schema

import "gorm.io/gorm"

type PlayerGame struct {
	gorm.Model
	UserID uint
	GameID uint
}
