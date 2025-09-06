package wrapper

import (
	"errors"
	"log/slog"
	"main/internal/domain/errs"
	viewErrs "main/internal/interfaces/http/html/errs"

	"github.com/gin-gonic/gin"
)

type ErrorHandler struct {
	viewErrorHandler  viewErrs.IErrorHandler
	logBusinessErrors bool
}

func NewErrorHandler(
	viewErrorHandler viewErrs.IErrorHandler,
	logBusinessErrors bool,
) ErrorHandler {
	return ErrorHandler{
		viewErrorHandler:  viewErrorHandler,
		logBusinessErrors: logBusinessErrors,
	}
}

func (e ErrorHandler) Handle(c *gin.Context, tmplName string, err error) {
	var bookingError *errs.BookingError

	if e.logBusinessErrors && errors.As(err, &bookingError) {
		slog.Info("BookingBusinessError",
			slog.Int("code", bookingError.Code),
			slog.String("message", bookingError.Message),
			slog.Any("classID", bookingError.ClassID),
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

	e.viewErrorHandler.Handle(c, tmplName, err)
}
