package wrapper

import (
	"errors"
	"log/slog"
	"main/internal/domain/errs"
	viewErrs "main/internal/interfaces/http/html/errs"

	"github.com/gin-gonic/gin"
)

type ErrorHandler struct {
	errorHandler      viewErrs.IErrorHandler
	logBusinessErrors bool
}

func NewErrorHandler(
	errorHandler viewErrs.IErrorHandler,
	logBusinessErrors bool,
) ErrorHandler {
	return ErrorHandler{
		errorHandler:      errorHandler,
		logBusinessErrors: logBusinessErrors,
	}
}

func (e ErrorHandler) Handle(c *gin.Context, tmplName string, err error) {
	var bookingError *errs.BookingError
	if e.logBusinessErrors && errors.As(err, &bookingError) {
		slog.Error("bookingBusinessError",
			slog.String("error", err.Error()),
			slog.Any("params", c.Request.URL.Query()),
			slog.String("endpoint", c.FullPath()),
		)
	}

	slog.Error("UnknownError",
		slog.String("error", err.Error()),
		slog.Any("params", c.Request.URL.Query()),
		slog.String("endpoint", c.FullPath()),
	)

	e.errorHandler.Handle(c, tmplName, err)
}
