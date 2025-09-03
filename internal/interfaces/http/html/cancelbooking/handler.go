package cancelbooking

import (
	"main/internal/domain/services"
	"main/internal/interfaces/http/err/handler"
	"main/internal/interfaces/http/html/dto"
	"net/http"

	"github.com/gin-gonic/gin"
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
	var form dto.BookingForm
	if err := c.ShouldBindQuery(&form); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	ctx := c.Request.Context()

	class, err := h.bookingService.CancelBooking(ctx, form.Token)
	if err != nil {
		h.errorHandler.HandleHTMLError(c, "err.tmpl", err)

		return
	}

	view, err := dto.ToBookingView(class)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	c.HTML(http.StatusOK, "confirmation_cancel_booking.tmpl", view)
}
