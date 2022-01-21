package containers

import (
	"fmt"
	"io"

	"github.com/bitterfly/go-chaos/hatgame/utils"
)

type Host struct {
	Players int
	Timer   int
}

func ParseHost(data io.ReadCloser) (*Host, error) {
	var container interface{} = &Host{}
	res, err := utils.Parse(data, container)
	if err != nil {
		return nil, err
	}

	host, ok := res.(*Host)
	if !ok {
		return nil, fmt.Errorf("could not convert to server User")
	}
	return host, nil
}
