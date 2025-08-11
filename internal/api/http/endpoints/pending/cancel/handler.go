package cancel

import (
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
	ctx := c.Request.Context()
	classIDStr := c.PostForm("classID")

	classID, err := uuid.Parse(classIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	cancelParams := models.CancelParams{
		ClassID:   classID,
		FirstName: c.PostForm("firstName"),
		Email:     c.PostForm("email"),
	}

	classID, err = h.ServicePendingOperations.CancelBooking(ctx, cancelParams)
	if err != nil {
		h.ErrorHandler.Handle(c, "err.tmpl", err)

		return
	}

	//TODO: different template?
	c.HTML(http.StatusOK, "submit.tmpl", gin.H{"ID": classID})
}
