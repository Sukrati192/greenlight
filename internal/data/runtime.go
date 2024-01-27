package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Runtime int32

var ErrRuntimeInvalidFormat = errors.New("invalid runtime format")

func (r Runtime) MarshalJSON() ([]byte, error) {
	jsonVal := fmt.Sprintf("%d mins", r)
	quotedJsonVal := strconv.Quote(jsonVal)
	return []byte(quotedJsonVal), nil
}

func (r *Runtime) UnmarshalJSON(val []byte) error {
	unquotedJSONValue, err := strconv.Unquote(string(val))
	if err != nil {
		return ErrRuntimeInvalidFormat
	}
	parts := strings.Split(unquotedJSONValue, " ")
	if len(parts) != 2 || parts[1] != "mins" {
		return ErrRuntimeInvalidFormat
	}
	duration, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return ErrRuntimeInvalidFormat
	}
	*r = Runtime(duration)
	return nil
}
