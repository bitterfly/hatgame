package schema

import (
	"encoding/json"
	"io"

	"gorm.io/gorm"
)

type Users struct {
	gorm.Model
	Email    string `gorm:"notnull"`
	Password uint   `gorm:"notnull"`
	Username string
	Avatar   []byte
}

func ParseUser(data io.ReadCloser) (*Users, error) {
	var user Users
	if err := json.NewDecoder(data).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}
