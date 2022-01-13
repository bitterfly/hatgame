package schema

import "gorm.io/gorm"

type Game struct {
	gorm.Model
	UserID      uint
	NumPlayers  int
	Timer       int
	NumWords    int
	PlayerWords []PlayerWord `gorm:"many2many:game_playerWords;"`
	Result      []Result     `gorm:"many2many:game_results;"`
}
