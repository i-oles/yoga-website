package create

import (
	"fmt"
	"main/internal/domain/services"
	"main/internal/errs"
	"net/http"

	"github.com/gin-gonic/gin"
)

// TODO: change to DTO
type Confirmation struct {
	//TODO: change names
	ClassType string `json:"class_type"`
	Date      string `json:"date"`
	Hour      string `json:"hour"`
	Place     string `json:"place"`
}

type Handler struct {
	serviceConfirmation services.IServiceConfirmation
	errorHandler        errs.ErrorHandler
}

func NewHandler(
	serviceConfirmation services.IServiceConfirmation,
	errorHandler errs.ErrorHandler,
) *Handler {
	return &Handler{
		serviceConfirmation: serviceConfirmation,
		errorHandler:        errorHandler,
	}
}

func (h *Handler) Handle(c *gin.Context) {
	ctx := c.Request.Context()
	confirmationToken := c.Query("token")

	class, err := h.serviceConfirmation.CreateBooking(ctx, confirmationToken)
	if err != nil {
		h.errorHandler.Handle(c, "err.tmpl", err)

		return
	}

	c.HTML(http.StatusOK, "confirmation.tmpl", gin.H{
		"date":       fmt.Sprintf("%d %s %d", class.StartTime.Day(), class.StartTime.Month(), class.StartTime.Year()),
		"hour":       fmt.Sprintf("%d:%02d", class.StartTime.Hour(), class.StartTime.Minute()),
		"place":      class.Location,
		"class_type": class.ClassCategory},
	)
}
