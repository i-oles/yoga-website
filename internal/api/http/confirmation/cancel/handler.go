package cancel

import (
	"fmt"
	errs2 "main/internal/domain/errs"
	"main/internal/domain/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// TODO: change to DTO
type confirmation struct {
	//TODO: change names
	ClassType string `json:"class_type"`
	Date      string `json:"date"`
	Hour      string `json:"hour"`
	Place     string `json:"place"`
}

type Handler struct {
	serviceConfirmation services.IServiceConfirmation
	errorHandler        errs2.ErrorHandler
}

func NewHandler(
	serviceConfirmation services.IServiceConfirmation,
	errorHandler errs2.ErrorHandler,
) *Handler {
	return &Handler{
		serviceConfirmation: serviceConfirmation,
		errorHandler:        errorHandler,
	}
}

func (h *Handler) Handle(c *gin.Context) {
	ctx := c.Request.Context()
	confirmationToken := c.Query("token")

	class, err := h.serviceConfirmation.CancelBooking(ctx, confirmationToken)
	if err != nil {
		h.errorHandler.Handle(c, "err.tmpl", err)

		return
	}

	//TODO: check this whole endpoint

	conf := confirmation{
		Date:      fmt.Sprintf("%d %s %d", class.StartTime.Day(), class.StartTime.Month(), class.StartTime.Year()),
		Hour:      fmt.Sprintf("%d:%02d", class.StartTime.Hour(), class.StartTime.Minute()),
		Place:     class.Location,
		ClassType: class.ClassCategory,
	}

	c.HTML(http.StatusOK, "cancellation.tmpl", conf)
}
