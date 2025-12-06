package wrapper

import (
	"log/slog"

	apiErrs "main/internal/interfaces/http/api/errs"

	"github.com/gin-gonic/gin"
)

type ErrorHandler struct {
	apiErrorHandler apiErrs.IErrorHandler
}

func NewErrorHandler(
	apiErrorHandler apiErrs.IErrorHandler,
) ErrorHandler {
	return ErrorHandler{
		apiErrorHandler: apiErrorHandler,
	}
}

func (e ErrorHandler) Handle(c *gin.Context, err error) {
	slog.Error("APIError",
		slog.String("error", err.Error()),
		slog.Any("params", c.Request.URL.Query()),
		slog.String("endpoint", c.FullPath()),
	)

	e.apiErrorHandler.Handle(c, err)
}
