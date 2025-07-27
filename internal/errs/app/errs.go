package app

import (
	"errors"
	"main/internal/errs"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorHandler struct{}

func NewErrorHandler() ErrorHandler {
	return ErrorHandler{}
}

func (e ErrorHandler) Handle(c *gin.Context, tmplName string, err error) {
	var bookingError *errs.BookingError
	switch {
	case errors.As(err, &bookingError):
		c.HTML(bookingError.Code, tmplName, gin.H{"error": bookingError.Message})
	default:
		c.HTML(http.StatusInternalServerError, tmplName, gin.H{"error": err.Error()})
	}
}
