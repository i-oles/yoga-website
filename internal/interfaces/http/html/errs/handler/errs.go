package viewErrHandler

import (
	"errors"
	"net/http"

	domainErrs "main/internal/domain/errs/view"

	"github.com/gin-gonic/gin"
)

type ErrorHandler struct{}

func NewErrorHandler() ErrorHandler {
	return ErrorHandler{}
}

func (e ErrorHandler) Handle(c *gin.Context, tmplName string, err error) {
	var viewError *domainErrs.ViewError
	if errors.As(err, &viewError) {
		switch viewError.Code {
		case domainErrs.BookingNotFoundCode:
			c.HTML(http.StatusNotFound, tmplName, gin.H{
				"ID":    viewError.ClassID,
				"Error": viewError.Message,
			})
		case domainErrs.BookingAlreadyExistsCode:
			c.HTML(http.StatusConflict, tmplName, gin.H{
				"ID":    viewError.ClassID,
				"Error": viewError.Message,
			})
		case domainErrs.ClassExpiredCode:
			c.HTML(http.StatusNotFound, tmplName, gin.H{
				"ID":    viewError.ClassID,
				"Error": viewError.Message,
			})
		case domainErrs.PendingBookingNotFoundCode:
			c.HTML(http.StatusNotFound, tmplName, gin.H{
				"Error": viewError.Message,
			})
		case domainErrs.TooManyPendingBookingsCode:
			c.HTML(http.StatusConflict, tmplName, gin.H{
				"ID":    viewError.ClassID,
				"Error": viewError.Message,
			})
		case domainErrs.ClassFullyBookedCode:
			c.HTML(http.StatusConflict, tmplName, gin.H{
				"ID":    viewError.ClassID,
				"Error": viewError.Message,
			})
		case domainErrs.SomeoneBookedClassFasterCode:
			c.HTML(http.StatusConflict, tmplName, gin.H{
				"Error": viewError.Message,
			})
		case domainErrs.ClassEmptyCode:
			c.HTML(http.StatusNotFound, tmplName, gin.H{
				"ID":    viewError.ClassID,
				"Error": viewError.Message,
			})
		case domainErrs.InvalidCancellationLinkCode:
			c.HTML(http.StatusNotFound, tmplName, gin.H{
				"Error": viewError.Message,
			})
		default:
			c.HTML(http.StatusInternalServerError, "err.tmpl", gin.H{
				"Error": "Coś poszło nie tak... Skontaktuj się ze mną :)",
			})
		}

		return
	}

	c.HTML(http.StatusInternalServerError, "err.tmpl", gin.H{
		"Error": "Coś poszło nie tak... Skontaktuj się ze mną :)",
	})
}
