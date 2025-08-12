package create

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
	ErrorHandler handler.IErrorHandler,
) *Handler {
	return &Handler{
		ServicePendingOperations: servicePendingOperations,
		ErrorHandler:             ErrorHandler,
	}
}

func (h *Handler) Handle(c *gin.Context) {
	var req dto.PendingOperationCreateRequest
	if err := c.ShouldBind(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	parsedUUID, err := uuid.Parse(req.ClassID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	createBookingParams := models.CreateParams{
		ClassID:   parsedUUID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
	}

	ctx := c.Request.Context()

	classID, err := h.ServicePendingOperations.CreateBooking(ctx, createBookingParams)
	if err != nil {
		h.ErrorHandler.Handle(c, "err.tmpl", err)

		return
	}

	resp := dto.PendingOperationCreateResponse{
		ClassID: classID,
	}

	c.HTML(http.StatusOK, "submit_create.tmpl", resp)
}
