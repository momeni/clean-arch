// Package serdser contains the reusable serialization/deserialization
// logics in order to be used by the resource packages.
package serdser

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/momeni/clean-arch/pkg/core/cerr"
)

// Bind tries to bind the request parameters, from the c context,
// into the req struct, received as an interface.
//
// The b indicates the binding method.
// Use binding.JSON in order to read json data from the body,
// binding.Query in order to read the URL query parameters, binding.Form
// in order to read a urlencoded or multipart form body for requests
// with a body and to read query parameters for GET requests,
// binding.Uri in order to read the path parameters.
func Bind(c *gin.Context, req any, b binding.Binding) bool {
	switch err := c.ShouldBindWith(req, b).(type) {
	case *validator.InvalidValidationError:
		c.JSON(http.StatusInternalServerError, gin.H{
			"detail": err.Error(),
		})
	case validator.ValidationErrors:
		var nameToErrs map[string][]string
		for _, ferr := range err {
			AddErr(&nameToErrs, ferr.Field(), ferr.Error())
		}
		c.JSON(http.StatusBadRequest, nameToErrs)
	default:
		if err == nil {
			return true
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"detail": err.Error(),
		})
	}
	return false
}

// AddErr adds the msgs error strings for the name field into the
// given errs map (instantiating it, if errs is nil yet).
func AddErr(errs *map[string][]string, name string, msgs ...string) {
	if (*errs) == nil {
		*errs = make(map[string][]string)
	}
	if elist, ok := (*errs)[name]; !ok {
		(*errs)[name] = msgs
	} else {
		(*errs)[name] = append(elist, msgs...)
	}
}

// Assert ensures that ok is true, and it was false, the name and msgs
// will be added to the errs map using the AddErr function.
func Assert(errs *map[string][]string, ok bool, name string, msgs ...string) bool {
	if ok {
		return true
	}
	AddErr(errs, name, msgs...)
	return false
}

// SerErr serializes the err error and transmits it as a JSON object
// with "detail" field containing the err string representation.
// If err is a *cerr.Error object, its HTTPStatusCode will be used for
// transmision of the error.
// Otherwise, a 500 response will be sent.
func SerErr(c *gin.Context, err error) {
	var ce *cerr.Error
	if errors.As(err, &ce) {
		c.JSON(ce.HTTPStatusCode, gin.H{
			"detail": ce.Err.Error(),
		})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{
		"detail": err.Error(),
	})
}
