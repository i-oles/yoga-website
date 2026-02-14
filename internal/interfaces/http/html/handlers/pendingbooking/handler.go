package pendingbooking

import (
	"net/http"
	"strings"

	"main/internal/domain/models"
	"main/internal/domain/services"
	"main/internal/interfaces/http/html/dto"
	viewErrs "main/internal/interfaces/http/html/errs"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type handler struct {
	PendingBookingsService services.IPendingBookingsService
	ViewErrorHandler       viewErrs.IErrorHandler
}

func NewHandler(
	pendingBookingsService services.IPendingBookingsService,
	viewErrorHandler viewErrs.IErrorHandler,
) *handler {
	return &handler{
		PendingBookingsService: pendingBookingsService,
		ViewErrorHandler:       viewErrorHandler,
	}
}

func (h *handler) Handle(c *gin.Context) {
	var form dto.PendingBookingForm
	if err := c.ShouldBind(&form); err != nil {
		viewErrs.ErrBadRequest(c, "pending_booking_form.tmpl", err)

		return
	}

	parsedUUID, err := uuid.Parse(form.ClassID)
	if err != nil {
		viewErrs.ErrBadRequest(c, "pending_booking_form.tmpl", err)

		return
	}

	pendingBookingParams := models.PendingBookingParams{
		ClassID:   parsedUUID,
		FirstName: form.FirstName,
		LastName:  form.LastName,
		Email:     strings.ToLower(form.Email),
	}

	ctx := c.Request.Context()

	classID, err := h.PendingBookingsService.CreatePendingBooking(ctx, pendingBookingParams)
	if err != nil {
		h.ViewErrorHandler.Handle(c, "pending_booking_form.tmpl", err)

		return
	}

	view := dto.PendingBookingView{
		ClassID: classID,
	}

	c.HTML(http.StatusOK, "pending_booking.tmpl", view)
}
