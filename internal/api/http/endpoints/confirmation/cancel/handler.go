package cancel

import (
	"main/internal/api/http/dto"
	"main/internal/api/http/err/handler"
	"main/internal/domain/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	serviceConfirmation services.IConfirmationService
	errorHandler        handler.IErrorHandler
}

func NewHandler(
	serviceConfirmation services.IConfirmationService,
	errorHandler handler.IErrorHandler,
) *Handler {
	return &Handler{
		serviceConfirmation: serviceConfirmation,
		errorHandler:        errorHandler,
	}
}

func (h *Handler) Handle(c *gin.Context) {
	var req dto.ConfirmationCancelRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	ctx := c.Request.Context()

	class, err := h.serviceConfirmation.CancelBooking(ctx, req.Token)
	if err != nil {
		h.errorHandler.Handle(c, "err.tmpl", err)

		return
	}

	c.HTML(http.StatusOK, "cancellation.tmpl", dto.ToConfirmationCancelResponse(class))
}
