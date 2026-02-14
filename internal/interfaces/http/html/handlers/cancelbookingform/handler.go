package cancelbookingform

import (
	"errors"
	"net/http"

	"main/internal/domain/services"
	"main/internal/interfaces/http/html/dto"
	viewErrs "main/internal/interfaces/http/html/errs"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type handler struct {
	bookingService   services.IBookingsService
	viewErrorHandler viewErrs.IErrorHandler
}

func NewHandler(
	bookingService services.IBookingsService,
	viewErrorHandler viewErrs.IErrorHandler,
) *handler {
	return &handler{
		bookingService:   bookingService,
		viewErrorHandler: viewErrorHandler,
	}
}

func (h *handler) Handle(c *gin.Context) {
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

	booking, err := h.bookingService.GetBookingForCancellation(ctx, bookingID, form.Token)
	if err != nil {
		h.viewErrorHandler.Handle(c, "err.tmpl", err)

		return
	}

	if booking.Class == nil {
		h.viewErrorHandler.Handle(c, "err.tmpl", errors.New("booking.Class should not be empty"))

		return
	}

	classView, err := dto.ToClassView(*booking.Class)
	if err != nil {
		viewErrs.ErrDTOConversion(c, "cancel_booking_form.tmpl", err)

		return
	}

	c.HTML(http.StatusOK, "cancel_booking_form.tmpl", gin.H{
		"Class": classView, "BookingID": bookingID, "ConfirmationToken": booking.ConfirmationToken,
	})
}
