package wrapper

import (
	"log/slog"
	"main/internal/interfaces/http/api/errs"

	"github.com/gin-gonic/gin"
)

type ErrorHandler struct {
	errorHandler apiErrs.IErrorHandler
}

func NewErrorHandler(
	errorHandler apiErrs.IErrorHandler,
) ErrorHandler {
	return ErrorHandler{
		errorHandler: errorHandler,
	}
}

func (e ErrorHandler) HandleJSONError(c *gin.Context, err error) {
	slog.Error("APIError",
		slog.String("error", err.Error()),
		slog.Any("params", c.Request.URL.Query()),
		slog.String("endpoint", c.FullPath()),
	)

	e.errorHandler.HandleJSONError(c, err)
}
