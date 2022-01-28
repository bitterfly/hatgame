package schema

import "gorm.io/gorm"

type Game struct {
	gorm.Model
	UserID     uint
	NumPlayers int
	Timer      int
	NumWords   int
	Result     []Result `gorm:"many2many:game_results;"`
}
