package schema

import "gorm.io/gorm"

type Result struct {
	gorm.Model
	TeamID uint
	Score  int
}
