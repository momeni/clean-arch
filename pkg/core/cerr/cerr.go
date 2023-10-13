// Package cerr represents the core layer errors.
// This package includes the Error struct which helps to wrap common
// errors with HTTPStatusCode, so the errors may be classified based
// on their types.
package cerr

import (
	"fmt"
	"net/http"
)

// Error represents an error, aka Err, and assigns a HTTPStatusCode
// http status code to that error based on its generic category.
type Error struct {
	Err            error
	HTTPStatusCode int
}

// Unwrap returns the wrapped inner error.
func (e *Error) Unwrap() error {
	return e.Err
}

// Error implements the error interface, returning a string
// representation of the Error instance.
func (e *Error) Error() string {
	return fmt.Sprintf("[%d] %s", e.HTTPStatusCode, e.Err.Error())
}

// BadRequest wraps the err error and marks it as a bad request, that
// is, the caller of the function which is returning this error is
// responsible for that error and may fix it by modifying the args
// of that function.
func BadRequest(err error) *Error {
	return &Error{Err: err, HTTPStatusCode: http.StatusBadRequest}
}

// Authentication wraps the err error and marks it as an authentication
// issue, that is, the caller is not identified and/or authenticated
// properly.
func Authentication(err error) *Error {
	return &Error{Err: err, HTTPStatusCode: http.StatusUnauthorized}
}

// Authorization wraps the err error and marks it as an authorization
// issue, that is, the caller is authenticated but does not have
// enough permission to invoke that function.
func Authorization(err error) *Error {
	return &Error{Err: err, HTTPStatusCode: http.StatusForbidden}
}

// NotFound wraps the err error and marks it as a not found issue, that
// is, the requested object does not exist.
func NotFound(err error) *Error {
	return &Error{Err: err, HTTPStatusCode: http.StatusNotFound}
}

// Conflict wraps the err error and marks it as a conflict issue, that
// is, the requested operation may not be accomplished due to the
// current conflicting system state.
func Conflict(err error) *Error {
	return &Error{Err: err, HTTPStatusCode: http.StatusConflict}
}
