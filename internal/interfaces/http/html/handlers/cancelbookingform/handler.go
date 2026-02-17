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

func (h *handler) Handle(ginCtx *gin.Context) {
	var uri dto.BookingCancelURI

	if err := ginCtx.ShouldBindUri(&uri); err != nil {
		viewErrs.ErrBadRequest(ginCtx, "cancel_booking_form.tmpl", err)

		return
	}

	var form dto.BookingCancelForm

	if err := ginCtx.ShouldBindQuery(&form); err != nil {
		viewErrs.ErrBadRequest(ginCtx, "cancel_booking_form.tmpl", err)

		return
	}

	bookingID, err := uuid.Parse(uri.BookingID)
	if err != nil {
		viewErrs.ErrBadRequest(ginCtx, "cancel_booking_form.tmpl", err)

		return
	}

	ctx := ginCtx.Request.Context()

	booking, err := h.bookingService.GetBookingForCancellation(ctx, bookingID, form.Token)
	if err != nil {
		h.viewErrorHandler.Handle(ginCtx, "err.tmpl", err)

		return
	}

	if booking.Class == nil {
		h.viewErrorHandler.Handle(ginCtx, "err.tmpl", errors.New("booking.Class should not be empty"))

		return
	}

	classView, err := dto.ToClassView(*booking.Class)
	if err != nil {
		viewErrs.ErrDTOConversion(ginCtx, "cancel_booking_form.tmpl", err)

		return
	}

	ginCtx.HTML(http.StatusOK, "cancel_booking_form.tmpl", gin.H{
		"Class": classView, "BookingID": bookingID, "ConfirmationToken": booking.ConfirmationToken,
	})
}
