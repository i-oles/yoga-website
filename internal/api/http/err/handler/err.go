package handler

import (
	"errors"
	domainErrs "main/internal/domain/errs"
	"net/http"

	"github.com/gin-gonic/gin"
)

type IErrorHandler interface {
	Handle(ctx *gin.Context, tmplName string, err error)
}

type ErrorHandler struct{}

func NewErrorHandler() ErrorHandler {
	return ErrorHandler{}
}

func (e ErrorHandler) Handle(c *gin.Context, tmplName string, err error) {
	var bookingError *domainErrs.BookingError
	switch {
	case errors.As(err, &bookingError):
		switch bookingError.Code {
		case domainErrs.ConfirmedBookingNotFoundCode:
			c.HTML(http.StatusOK, tmplName, gin.H{
				"ID":    bookingError.ClassID,
				"Error": bookingError.Message})
		}
	default:
		c.HTML(http.StatusInternalServerError, tmplName, gin.H{"error": err.Error()})
	}
}
