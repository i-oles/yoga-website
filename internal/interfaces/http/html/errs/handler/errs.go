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
	var businessError *domainErrs.BusinessError
	if errors.As(err, &businessError) {
		switch businessError.Code {
		case domainErrs.BookingNotFoundCode,
			domainErrs.ClassExpiredCode,
			domainErrs.ClassEmptyCode:
			ctx.HTML(http.StatusNotFound, tmplName, gin.H{
				"ID":    businessError.ClassID,
				"Error": businessError.Message,
			})

			return
		case domainErrs.BookingAlreadyExistsCode,
			domainErrs.TooManyPendingBookingsCode,
			domainErrs.ClassFullyBookedCode:
			ctx.HTML(http.StatusConflict, tmplName, gin.H{
				"ID":    businessError.ClassID,
				"Error": businessError.Message,
			})

			return
		case domainErrs.PendingBookingNotFoundCode,
			domainErrs.InvalidCancellationLinkCode:
			ctx.HTML(http.StatusNotFound, tmplName, gin.H{
				"Error": businessError.Message,
			})

			return
		case domainErrs.SomeoneBookedClassFasterCode:
			ctx.HTML(http.StatusConflict, tmplName, gin.H{
				"Error": businessError.Message,
			})

			return
		default:
			ctx.HTML(http.StatusInternalServerError, "err.tmpl", gin.H{
				"Error": "error_id: " + ctx.GetString("request_id"),
			})
		}
	}

	e.handleInternalError(ctx)
}

func (e errorHandler) handleInternalError(ctx *gin.Context) {
	if ctx.GetHeader("HX-Request") == "true" {
		ctx.Header("HX-Redirect", "/error")
		ctx.Status(http.StatusInternalServerError)

		return
	}

	ctx.HTML(http.StatusInternalServerError, "err.tmpl", gin.H{
		"Error": "error_id: " + ctx.GetString("request_id"),
	})
}
