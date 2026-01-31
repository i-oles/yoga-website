package wrapper

import (
	"errors"
	"log/slog"

	domainErrs "main/internal/domain/errs/view"
	handlerErrs "main/internal/interfaces/http/html/errs"

	"github.com/gin-gonic/gin"
)

type ErrorHandler struct {
	errorHandler      handlerErrs.IErrorHandler
	logBusinessErrors bool
}

func NewErrorHandler(
	errorHandler handlerErrs.IErrorHandler,
	logBusinessErrors bool,
) ErrorHandler {
	return ErrorHandler{
		errorHandler:      errorHandler,
		logBusinessErrors: logBusinessErrors,
	}
}

func (e ErrorHandler) Handle(c *gin.Context, tmplName string, err error) {
	var viewError *domainErrs.ViewError

	if e.logBusinessErrors && errors.As(err, &viewError) {
		slog.Info("BookingBusinessError",
			slog.Int("code", viewError.Code),
			slog.String("message", viewError.Message),
			slog.Any("classID", viewError.ClassID),
			slog.Any("params", c.Request.URL.Query()),
			slog.String("endpoint", c.FullPath()),
		)
	} else {
		slog.Error("UnknownError",
			slog.String("error", err.Error()),
			slog.Any("params", c.Request.URL.Query()),
			slog.String("endpoint", c.FullPath()),
		)
	}

	e.errorHandler.Handle(c, tmplName, err)
}
