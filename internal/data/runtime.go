package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Runtime int32

// MarshalJSON encodes json
func (r Runtime) MarshalJSON() ([]byte, error) {
	value := strconv.Quote(fmt.Sprintf("%d mins", r))

	return []byte(value), nil
}

// UnmarshalJSON decodes json
func (r *Runtime) UnmarshalJSON(value []byte) error {
	unmarshalValue, err := strconv.Unquote(string(value))
	if err != nil {
		return err
	}

	parts := strings.Split(unmarshalValue, " ")
	if len(parts) != 1 && parts[1] != "mins" {
		return errors.New("incorrect runtime value format")
	}

	i, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return err
	}

	*r = Runtime(i)

	return nil
}
