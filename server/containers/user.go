package containers

import (
	"fmt"
	"io"

	"github.com/bitterfly/go-chaos/hatgame/utils"
)

type User struct {
	Email    string
	Password string
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
