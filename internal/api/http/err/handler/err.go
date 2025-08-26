package handler

import (
	"errors"
	domainErrs "main/internal/domain/errs"
	"net/http"

	"github.com/gin-gonic/gin"
)

type IErrorHandler interface {
	HandleHTMLError(ctx *gin.Context, tmplName string, err error)
	HandleJSONError(ctx *gin.Context, err error)
}

type ErrorHandler struct{}

func NewErrorHandler() ErrorHandler {
	return ErrorHandler{}
}

func (e ErrorHandler) HandleHTMLError(c *gin.Context, tmplName string, err error) {
	var bookingError *domainErrs.BookingError
	if errors.As(err, &bookingError) {
		switch bookingError.Code {
		case domainErrs.ConfirmedBookingNotFoundCode:
			c.HTML(http.StatusNotFound, tmplName, gin.H{
				"ID":    bookingError.ClassID,
				"Error": bookingError.Message,
			})
		case domainErrs.ConfirmedBookingAlreadyExistsCode:
			c.HTML(http.StatusConflict, tmplName, gin.H{
				"Error": bookingError.Message,
			})
		case domainErrs.ExpiredClassBookingCode:
			c.HTML(http.StatusNotFound, tmplName, gin.H{
				"ID":    bookingError.ClassID,
				"Error": bookingError.Message,
			})
		case domainErrs.PendingOperationNotFoundCode:
			c.HTML(http.StatusNotFound, tmplName, gin.H{
				"Error": bookingError.Message,
			})
		case domainErrs.TooManyPendingOperationsCode:
			c.HTML(http.StatusConflict, tmplName, gin.H{
				"ID":    bookingError.ClassID,
				"Error": bookingError.Message,
			})
		case domainErrs.ClassFullyBookedCode:
			c.HTML(http.StatusConflict, tmplName, gin.H{
				"ID":    bookingError.ClassID,
				"Error": bookingError.Message,
			})
		case domainErrs.SomeoneBookedClassFasterCode:
			c.HTML(http.StatusConflict, tmplName, gin.H{
				"Error": bookingError.Message,
			})
		case domainErrs.ClassEmptyCode:
			c.HTML(http.StatusNotFound, tmplName, gin.H{
				"ID":    bookingError.ClassID,
				"Error": bookingError.Message,
			})
		default:
			c.HTML(http.StatusInternalServerError, "err.tmpl", gin.H{
				"Error": "Coś poszło nie tak... Contact me :)",
			})
		}

		return
	}

	c.HTML(http.StatusInternalServerError, "err.tmpl", gin.H{
		"Error": "Coś poszło nie tak... Contact me :)",
	})

	return
}

func (e ErrorHandler) HandleJSONError(c *gin.Context, err error) {
	var classError *domainErrs.ClassError
	if errors.As(err, &classError) {
		switch classError.Code {
		case domainErrs.BadRequestCode:
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		return
	}

	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

	return
}
