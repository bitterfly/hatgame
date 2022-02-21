package schema

import "gorm.io/gorm"

type GameWord struct {
	gorm.Model
	PlayerWordID   uint
	GuessedByID    uint
	GameID         uint
	UserDictionary UserDictionary `gorm:"foreignKey:PlayerWordID"`
	GuessedBy      User           `gorm:"foreignKey:GuessedByID"`
	Game           Game           `gorm:"foreignKey:GameID"`
}
