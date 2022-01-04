package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
)

func Parse(data io.ReadCloser, container interface{}) (interface{}, error) {
	if err := json.NewDecoder(data).Decode(container); err != nil {
		return nil, err
	}
	return container, nil
}

func ParseUint(vars map[string]string, name string) (uint, error) {
	valueS, ok := vars[name]
	if !ok {
		return 0, fmt.Errorf("missing parameter for id")
	}
	value, err := strconv.ParseUint(valueS, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("could not parse %s as uint", valueS)
	}
	return uint(value), nil
}

func ParseInt(vars map[string]string, name string) (int, error) {
	valueS, ok := vars[name]
	if !ok {
		return 0, fmt.Errorf("missing parameter for id")
	}
	value, err := strconv.ParseInt(valueS, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("could not parse %s as int", valueS)
	}
	return int(value), nil
}
