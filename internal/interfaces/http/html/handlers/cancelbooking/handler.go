package cancelbooking

import (
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

	err = h.bookingService.CancelBooking(ctx, bookingID, form.Token)
	if err != nil {
		h.viewErrorHandler.Handle(ginCtx, "cancel_booking_form.tmpl", err)

		return
	}

	ginCtx.HTML(http.StatusOK, "confirmation_cancel_booking.tmpl", gin.H{})
}
