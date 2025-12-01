package cancelbooking

import (
	"net/http"

	"main/internal/domain/services"
	"main/internal/interfaces/http/html/dto"
	viewErrs "main/internal/interfaces/http/html/errs"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	bookingService   services.IBookingsService
	viewErrorHandler viewErrs.IErrorHandler
}

func NewHandler(
	bookingService services.IBookingsService,
	viewErrorHandler viewErrs.IErrorHandler,
) *Handler {
	return &Handler{
		bookingService:   bookingService,
		viewErrorHandler: viewErrorHandler,
	}
}

func (h *Handler) Handle(c *gin.Context) {
	var uri dto.BookingCancelURI

	if err := c.ShouldBindUri(&uri); err != nil {
		viewErrs.ErrBadRequest(c, "cancel_booking_form.tmpl", err)

		return
	}

	var form dto.BookingCancelForm

	if err := c.ShouldBindQuery(&form); err != nil {
		viewErrs.ErrBadRequest(c, "cancel_booking_form.tmpl", err)

		return
	}

	bookingID, err := uuid.Parse(uri.BookingID)
	if err != nil {
		viewErrs.ErrBadRequest(c, "cancel_booking_form.tmpl", err)

		return
	}

	ctx := c.Request.Context()

	err = h.bookingService.CancelBooking(ctx, bookingID, form.Token)
	if err != nil {
		h.viewErrorHandler.Handle(c, "cancel_booking_form.tmpl", err)

		return
	}

	c.HTML(http.StatusOK, "confirmation_cancel_booking.tmpl", gin.H{})
}
