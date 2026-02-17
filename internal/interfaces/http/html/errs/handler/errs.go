package viewErrHandler

import (
	"errors"
	"net/http"

	domainErrs "main/internal/domain/errs/view"

	"github.com/gin-gonic/gin"
)

type errorHandler struct{}

func NewErrorHandler() errorHandler {
	return errorHandler{}
}

func (e errorHandler) Handle(ctx *gin.Context, tmplName string, err error) {
	var viewError *domainErrs.ViewError
	if errors.As(err, &viewError) {
		switch viewError.Code {
		case domainErrs.BookingNotFoundCode,
			domainErrs.ClassExpiredCode,
			domainErrs.ClassEmptyCode:
			ctx.HTML(http.StatusNotFound, tmplName, gin.H{
				"ID":    viewError.ClassID,
				"Error": viewError.Message,
			})
		case domainErrs.BookingAlreadyExistsCode,
			domainErrs.TooManyPendingBookingsCode,
			domainErrs.ClassFullyBookedCode:
			ctx.HTML(http.StatusConflict, tmplName, gin.H{
				"ID":    viewError.ClassID,
				"Error": viewError.Message,
			})
		case domainErrs.PendingBookingNotFoundCode,
			domainErrs.InvalidCancellationLinkCode:
			ctx.HTML(http.StatusNotFound, tmplName, gin.H{
				"Error": viewError.Message,
			})
		case domainErrs.SomeoneBookedClassFasterCode:
			ctx.HTML(http.StatusConflict, tmplName, gin.H{
				"Error": viewError.Message,
			})
		default:
			ctx.HTML(http.StatusInternalServerError, "err.tmpl", gin.H{
				"Error": "Coś poszło nie tak... Skontaktuj się ze mną :)",
			})
		}

		return
	}

	ctx.HTML(http.StatusInternalServerError, "err.tmpl", gin.H{
		"Error": "Coś poszło nie tak... Skontaktuj się ze mną :)",
	})
}
