package httperr

import (
	"fmt"
)

type Error struct {
	Message    string
	StatusCode int
	Err        error
}

func New(statusCode int, format string, a ...interface{}) *Error {
	return &Error{
		Message:    fmt.Sprintf(format, a...),
		StatusCode: statusCode,
	}
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err.Error())
	}
	return e.Message
}

func (e *Error) Wrap(err error) *Error {
	e.Err = err
	return e
}

func Wrap(err error, statusCode int, format string, a ...interface{}) *Error {
	return New(statusCode, format, a...).Wrap(err)
}
