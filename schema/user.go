package schema

import (
	"fmt"
	"io"

	"github.com/bitterfly/go-chaos/hatgame/utils"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email    string `gorm:"notnull"`
	Password []byte `gorm:"notnull"`
	Username string
	Avatar   []byte
}

func ParseUser(data io.ReadCloser) (*User, error) {
	var container interface{} = &User{}
	res, err := utils.Parse(data, container)
	if err != nil {
		return nil, err
	}

	user, ok := res.(*User)
	if !ok {
		return nil, fmt.Errorf("could not convert to server User")
	}
	return user, nil
}
