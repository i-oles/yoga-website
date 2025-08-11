package log

import (
	"log/slog"
	"main/internal/api/http/err"

	"github.com/gin-gonic/gin"
)

type ErrorHandler struct {
	errorHandler err.ErrorHandler
}

func NewErrorHandler(
	errorHandler err.ErrorHandler,
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
