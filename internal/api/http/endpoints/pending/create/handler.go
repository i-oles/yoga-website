package create

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
	ErrorHandler handler.IErrorHandler,
) *Handler {
	return &Handler{
		ServicePendingOperations: servicePendingOperations,
		ErrorHandler:             ErrorHandler,
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

	firstName := c.PostForm("first_name")
	lastName := c.PostForm("last_name")
	email := c.PostForm("email")

	submitBooking := models.CreateParams{
		ClassID:   classID,
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
	}

	classID, err = h.ServicePendingOperations.CreateBooking(ctx, submitBooking)
	if err != nil {
		h.ErrorHandler.Handle(c, "err.tmpl", err)

		return
	}

	c.HTML(http.StatusOK, "submit.tmpl", gin.H{"ID": classID})
}
