package utils

import (
	"encoding/json"
	"io"
)

func Parse(data io.ReadCloser, container interface{}) (interface{}, error) {
	if err := json.NewDecoder(data).Decode(container); err != nil {
		return nil, err
	}
	return container, nil
}
