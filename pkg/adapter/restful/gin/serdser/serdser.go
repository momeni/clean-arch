package serdser

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/momeni/clean-arch/pkg/core/cerr"
)

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

func Assert(errs *map[string][]string, ok bool, name string, msgs ...string) bool {
	if ok {
		return true
	}
	AddErr(errs, name, msgs...)
	return false
}

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
