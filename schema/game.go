package schema

import "gorm.io/gorm"

type Game struct {
	gorm.Model
	UserID      uint
	NumPlayers  int
	Timer       int
	NumWords    int
	PlayerWords []PlayerWord `gorm:"many2many:game_playerWords;"`
	Teams       []Team       `gorm:"many2many:game_teams;"`
}
