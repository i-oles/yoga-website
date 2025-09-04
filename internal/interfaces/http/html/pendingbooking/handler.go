package pendingbooking

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
	PendingBookingsService services.IPendingBookingsService
	ErrorHandler           handler.IErrorHandler
}

func NewHandler(
	pendingBookingsService services.IPendingBookingsService,
	errorHandler handler.IErrorHandler,
) *Handler {
	return &Handler{
		PendingBookingsService: pendingBookingsService,
		ErrorHandler:           errorHandler,
	}
}

func (h *Handler) Handle(c *gin.Context) {
	var form dto.PendingBookingForm
	if err := c.ShouldBind(&form); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	parsedUUID, err := uuid.Parse(form.ClassID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	pendingBookingParams := models.PendingBookingParams{
		ClassID:   parsedUUID,
		FirstName: form.FirstName,
		LastName:  form.LastName,
		Email:     strings.ToLower(form.Email),
	}

	ctx := c.Request.Context()

	classID, err := h.PendingBookingsService.CreatePendingBooking(ctx, pendingBookingParams)
	if err != nil {
		h.ErrorHandler.HandleHTMLError(c, "bookings_pending_form.tmpl", err)

		return
	}

	view := dto.PendingBookingView{
		ClassID: classID,
	}

	c.HTML(http.StatusOK, "pending_booking.tmpl", view)
}
