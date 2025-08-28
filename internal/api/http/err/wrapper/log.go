package wrapper

import (
	"log/slog"
	"main/internal/api/http/err/handler"

	"github.com/gin-gonic/gin"
)

type ErrorHandler struct {
	errorHandler      handler.IErrorHandler
	logBusinessErrors bool
}

func NewErrorHandler(
	errorHandler handler.IErrorHandler,
	logBusinessErrors bool,
) ErrorHandler {
	return ErrorHandler{
		errorHandler:      errorHandler,
		logBusinessErrors: logBusinessErrors,
	}
}

func (e ErrorHandler) HandleHTMLError(c *gin.Context, tmplName string, err error) {
	if e.logBusinessErrors {
		slog.Error("bookingBusinessError",
			slog.String("error", err.Error()),
			slog.Any("params", c.Request.URL.Query()),
			slog.String("endpoint", c.FullPath()),
		)
	}

	e.errorHandler.HandleHTMLError(c, tmplName, err)
}

func (e ErrorHandler) HandleJSONError(c *gin.Context, err error) {
	slog.Error("classOwnerError",
		slog.String("error", err.Error()),
		slog.Any("params", c.Request.URL.Query()),
		slog.String("endpoint", c.FullPath()),
	)

	e.errorHandler.HandleJSONError(c, err)
}
