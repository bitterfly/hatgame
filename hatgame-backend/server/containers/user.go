package containers

import (
	"fmt"
	"io"

	"github.com/bitterfly/go-chaos/hatgame/utils"
)

type LoginUser struct {
	Email    string
	Password string
	Username string
}

func ParseLoginUser(data io.ReadCloser) (*LoginUser, error) {
	var container interface{} = &LoginUser{}
	res, err := utils.Parse(data, container)
	if err != nil {
		return nil, err
	}

	user, ok := res.(*LoginUser)
	if !ok {
		return nil, fmt.Errorf("could not convert to server User")
	}
	return user, nil
}

type User struct {
	ID       uint
	Email    string
	Username string
}
