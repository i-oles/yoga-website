package cancelbooking

import (
	"main/internal/domain/services"
	"main/internal/interfaces/http/err/handler"
	"main/internal/interfaces/http/html/dto"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	bookingService services.IBookingsService
	errorHandler   handler.IErrorHandler
}

func NewHandler(
	bookingService services.IBookingsService,
	errorHandler handler.IErrorHandler,
) *Handler {
	return &Handler{
		bookingService: bookingService,
		errorHandler:   errorHandler,
	}
}

func (h *Handler) Handle(c *gin.Context) {
	var form dto.BookingCancelForm

	if err := c.ShouldBindUri(&form); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	if err := c.ShouldBindQuery(&form); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	bookingID, err := uuid.Parse(form.BookingID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	ctx := c.Request.Context()

	err = h.bookingService.CancelBooking(ctx, bookingID, form.Token)
	if err != nil {
		h.errorHandler.HandleHTMLError(c, "cancel_booking_form.tmpl", err)

		return
	}

	c.HTML(http.StatusOK, "confirmation_cancel_booking.tmpl", gin.H{})
}
