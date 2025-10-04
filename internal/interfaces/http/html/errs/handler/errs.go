package viewErrHandler

import (
	"errors"
	domainErrs "main/internal/domain/errs"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorHandler struct{}

func NewErrorHandler() ErrorHandler {
	return ErrorHandler{}
}

func (e ErrorHandler) Handle(c *gin.Context, tmplName string, err error) {
	var bookingError *domainErrs.BookingError
	if errors.As(err, &bookingError) {
		switch bookingError.Code {
		case domainErrs.BookingNotFoundCode:
			c.HTML(http.StatusNotFound, tmplName, gin.H{
				"ID":    bookingError.ClassID,
				"Error": bookingError.Message,
			})
		case domainErrs.BookingAlreadyExistsCode:
			c.HTML(http.StatusConflict, tmplName, gin.H{
				"ID":    bookingError.ClassID,
				"Error": bookingError.Message,
			})
		case domainErrs.ClassExpiredCode:
			c.HTML(http.StatusNotFound, tmplName, gin.H{
				"ID":    bookingError.ClassID,
				"Error": bookingError.Message,
			})
		case domainErrs.PendingBookingNotFoundCode:
			c.HTML(http.StatusNotFound, tmplName, gin.H{
				"Error": bookingError.Message,
			})
		case domainErrs.TooManyPendingBookingsCode:
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
		case domainErrs.InvalidCancellationLinkCode:
			c.HTML(http.StatusNotFound, tmplName, gin.H{
				"Error": bookingError.Message,
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

	return
}
