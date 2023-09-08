package cerr

import (
	"fmt"
	"net/http"
)

type Error struct {
	Err            error
	HTTPStatusCode int
}

func (e *Error) Unwrap() error {
	return e.Err
}

func (e *Error) Error() string {
	return fmt.Sprintf("[%d] %s", e.HTTPStatusCode, e.Err.Error())
}

func BadRequest(err error) *Error {
	return &Error{Err: err, HTTPStatusCode: http.StatusBadRequest}
}

func Authentication(err error) *Error {
	return &Error{Err: err, HTTPStatusCode: http.StatusUnauthorized}
}

func Authorization(err error) *Error {
	return &Error{Err: err, HTTPStatusCode: http.StatusForbidden}
}

func NotFound(err error) *Error {
	return &Error{Err: err, HTTPStatusCode: http.StatusNotFound}
}

func Conflict(err error) *Error {
	return &Error{Err: err, HTTPStatusCode: http.StatusConflict}
}
