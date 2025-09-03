package cancelbooking

import (
	"main/internal/domain/models"
	"main/internal/domain/services"
	"main/internal/interfaces/http/err/handler"
	"main/internal/interfaces/http/html/dto"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	ServicePendingOperations services.IPendingBookingsService
	ErrorHandler             handler.IErrorHandler
}

func NewHandler(
	servicePendingOperations services.IPendingBookingsService,
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

	cancelParams := models.CancelBookingParams{
		ClassID: parsedUUID,
		Email:   strings.ToLower(req.Email),
	}

	ctx := c.Request.Context()

	classID, err := h.ServicePendingOperations.CancelPendingBooking(ctx, cancelParams)
	if err != nil {
		h.ErrorHandler.HandleHTMLError(c, "cancel.tmpl", err)

		return
	}

	resp := dto.PendingOperationCancelResponse{
		ClassID: classID,
	}

	c.HTML(http.StatusOK, "pending_operation_cancel.tmpl", resp)
}
