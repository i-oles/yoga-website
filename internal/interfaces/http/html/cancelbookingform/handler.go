package cancelbookingform

import (
	"main/internal/domain/services"
	"main/internal/interfaces/http/err/handler"
	"main/internal/interfaces/http/html/dto"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	BookingService services.IBookingsService
	ErrorHandler   handler.IErrorHandler
}

func NewHandler(
	bookingService services.IBookingsService,
	errorHandler handler.IErrorHandler,
) *Handler {
	return &Handler{
		BookingService: bookingService,
		ErrorHandler:   errorHandler,
	}
}

func (h *Handler) Handle(c *gin.Context) {
	var form dto.BookingCancelForm

	if err := c.ShouldBindQuery(&form); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	if err := c.ShouldBindUri(&form); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	bookingID, err := uuid.Parse(form.BookingID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	ctx := c.Request.Context()

	cancelledBooking, err := h.BookingService.CancelBookingForm(ctx, bookingID, form.Token)
	if err != nil {
		h.ErrorHandler.HandleHTMLError(c, "cancel_booking_form.tmpl", err)

		return
	}

	view, err := dto.ToBookingCancelView(cancelledBooking)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	c.HTML(http.StatusOK, "cancel_booking_form.tmpl", view)
}
