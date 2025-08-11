package wrapper

import (
	"log/slog"
	"main/internal/api/http/err/handler"

	"github.com/gin-gonic/gin"
)

type ErrorHandler struct {
	errorHandler handler.IErrorHandler
}

func NewErrorHandler(
	errorHandler handler.IErrorHandler,
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
