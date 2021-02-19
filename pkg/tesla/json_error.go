package tesla

import (
	"errors"
	"fmt"
)

// ErrUnauthorized is when we failed to log into the gateway
var ErrUnauthorized = errors.New("User does not have adequate access rights")

type jsonError struct {
	Code    int    `json:"code"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

func (j jsonError) getError() error {
	if j.Code == 403 && j.Message == ErrUnauthorized.Error() {
		return ErrUnauthorized
	}
	if j.Code == 0 && j.Error == "" || j.Message == "" {
		return nil
	}
	return fmt.Errorf("error from api: code %d, error: %q, message: %q",
		j.Code,
		j.Error,
		j.Message,
	)
}
