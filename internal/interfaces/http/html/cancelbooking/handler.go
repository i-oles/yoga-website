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
	var uri dto.BookingCancelURI

	if err := c.ShouldBindUri(&uri); err != nil {
		h.errorHandler.HandleHTMLError(c, "cancel_booking_form.tmpl", err)
		return
	}

	var form dto.BookingCancelForm

	if err := c.ShouldBindQuery(&form); err != nil {
		h.errorHandler.HandleHTMLError(c, "cancel_booking_form.tmpl", err)
		return
	}

	bookingID, err := uuid.Parse(uri.BookingID)
	if err != nil {
		h.errorHandler.HandleHTMLError(c, "cancel_booking_form.tmpl", err)
		return
	}

	ctx := c.Request.Context()

	err = h.bookingService.CancelBooking(ctx, bookingID, form.Token)
	if err != nil {
		h.errorHandler.HandleHTMLError(c, "cancel_booking_form.tmpl", err)
		return
	}

	c.HTML(http.StatusOK, "confirmation_cancel_booking.tmpl", gin.H{})
}
