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

func (e ErrorHandler) Handle(ctx *gin.Context, err error) {
	slog.Error("APIError",
		slog.String("error", err.Error()),
		slog.Any("params", ctx.Request.URL.Query()),
		slog.String("endpoint", ctx.FullPath()),
	)

	e.apiErrorHandler.Handle(ctx, err)
}
