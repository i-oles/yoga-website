package log

import (
	"log/slog"
	"main/internal/errs"

	"github.com/gin-gonic/gin"
)

type ErrorHandler struct {
	errorHandler errs.ErrorHandler
}

func NewErrorHandler(
	errorHandler errs.ErrorHandler,
) ErrorHandler {
	return ErrorHandler{
		errorHandler: errorHandler,
	}
}

func (e ErrorHandler) Handle(c *gin.Context, tmplName string, err error) {
	slog.Error("",
		slog.String("error", err.Error()),
		slog.Any("params", c.Request.URL.Query()),
		slog.String("endpoint", c.FullPath()),
	)

	e.errorHandler.Handle(c, tmplName, err)
}
