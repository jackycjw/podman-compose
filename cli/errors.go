package cli

import (
	"encoding/json"
	"github.com/pkg/errors"
	"io"
)

var (
	ErrNotImplemented = errors.New("function not implemented")
)

type ErrorModel struct {
	// API root cause formatted for automated parsing
	// example: API root cause
	Because string `json:"cause"`
	// human error message, formatted for a human to read
	// example: human error message
	Message string `json:"message"`
	// http response code
	ResponseCode int `json:"response"`
}

func (e ErrorModel) Error() string {
	return e.Message
}

func (e ErrorModel) Cause() error {
	return errors.New(e.Because)
}

func (e ErrorModel) Code() int {
	return e.ResponseCode
}

func handleError(data []byte) error {
	e := ErrorModel{}
	if err := json.Unmarshal(data, &e); err != nil {
		return err
	}
	return e
}

func (a APIResponse) Process(unmarshalInto interface{}) error {
	data, err := io.ReadAll(a.Response.Body)
	if err != nil {
		return errors.Wrap(err, "unable to process API response")
	}
	if a.IsSuccess() || a.IsRedirection() {
		if unmarshalInto != nil {
			return json.Unmarshal(data, unmarshalInto)
		}
		return nil
	}
	// TODO should we add a debug here with the response code?
	return handleError(data)
}

func CheckResponseCode(inError error) (int, error) {
	e, ok := inError.(ErrorModel)
	if !ok {
		return -1, errors.New("error is not type ErrorModel")
	}
	return e.Code(), nil
}
