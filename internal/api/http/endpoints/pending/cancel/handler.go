package cancel

import (
	"main/internal/api/http/dto"
	"main/internal/api/http/err/handler"
	"main/internal/domain/models"
	"main/internal/domain/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	ServicePendingOperations services.IPendingOperationsService
	ErrorHandler             handler.IErrorHandler
}

func NewHandler(
	servicePendingOperations services.IPendingOperationsService,
	errorHandler handler.IErrorHandler,
) *Handler {
	return &Handler{
		ServicePendingOperations: servicePendingOperations,
		ErrorHandler:             errorHandler,
	}
}

func (h *Handler) Handle(c *gin.Context) {
	var req dto.PendingOperationCancelRequest
	if err := c.ShouldBind(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	parsedUUID, err := uuid.Parse(req.ClassID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	cancelParams := models.CancelParams{
		ClassID: parsedUUID,
		Email:   req.Email,
	}

	ctx := c.Request.Context()

	classID, err := h.ServicePendingOperations.CancelBooking(ctx, cancelParams)
	if err != nil {
		h.ErrorHandler.Handle(c, "cancel.tmpl", err)

		return
	}

	resp := dto.PendingOperationCancelResponse{
		ClassID: classID,
	}

	c.HTML(http.StatusOK, "pending_operation_cancel.tmpl", resp)
}
