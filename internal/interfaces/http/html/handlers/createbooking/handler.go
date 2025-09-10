package createbooking

import (
	"main/internal/domain/services"
	"main/internal/interfaces/http/html/dto"
	viewErrs "main/internal/interfaces/http/html/errs"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	bookingsService  services.IBookingsService
	viewErrorHandler viewErrs.IErrorHandler
}

func NewHandler(
	bookingService services.IBookingsService,
	viewErrorHandler viewErrs.IErrorHandler,
) *Handler {
	return &Handler{
		bookingsService:  bookingService,
		viewErrorHandler: viewErrorHandler,
	}
}

func (h *Handler) Handle(c *gin.Context) {
	var form dto.BookingCreateForm
	if err := c.ShouldBindQuery(&form); err != nil {
		viewErrs.ErrBadRequest(c, "err.tmpl", err)

		return
	}

	ctx := c.Request.Context()

	class, err := h.bookingsService.CreateBooking(ctx, form.Token)
	if err != nil {
		h.viewErrorHandler.Handle(c, "err.tmpl", err)

		return
	}

	view, err := dto.ToBookingView(class)
	if err != nil {
		viewErrs.ErrDTOConversion(c, "err.tmpl", err)

		return
	}

	c.HTML(http.StatusOK, "confirmation_create_booking.tmpl", view)
}
