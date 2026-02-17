package createbooking

import (
	"net/http"

	"main/internal/domain/services"
	"main/internal/interfaces/http/html/dto"
	viewErrs "main/internal/interfaces/http/html/errs"

	"github.com/gin-gonic/gin"
)

type handler struct {
	bookingsService  services.IBookingsService
	viewErrorHandler viewErrs.IErrorHandler
}

func NewHandler(
	bookingService services.IBookingsService,
	viewErrorHandler viewErrs.IErrorHandler,
) *handler {
	return &handler{
		bookingsService:  bookingService,
		viewErrorHandler: viewErrorHandler,
	}
}

func (h *handler) Handle(ginCtx *gin.Context) {
	var form dto.BookingCreateForm
	if err := ginCtx.ShouldBindQuery(&form); err != nil {
		viewErrs.ErrBadRequest(ginCtx, "err.tmpl", err)

		return
	}

	ctx := ginCtx.Request.Context()

	class, err := h.bookingsService.CreateBooking(ctx, form.Token)
	if err != nil {
		h.viewErrorHandler.Handle(ginCtx, "err.tmpl", err)

		return
	}

	view, err := dto.ToClassView(class)
	if err != nil {
		viewErrs.ErrDTOConversion(ginCtx, "err.tmpl", err)

		return
	}

	ginCtx.HTML(http.StatusOK, "confirmation_create_booking.tmpl", view)
}
