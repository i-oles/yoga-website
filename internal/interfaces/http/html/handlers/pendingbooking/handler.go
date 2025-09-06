package pendingbooking

import (
	"fmt"
	"main/internal/domain/models"
	"main/internal/domain/services"
	"main/internal/interfaces/http/html/dto"
	viewErrs "main/internal/interfaces/http/html/errs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	PendingBookingsService services.IPendingBookingsService
	ViewErrorHandler       viewErrs.IErrorHandler
}

func NewHandler(
	pendingBookingsService services.IPendingBookingsService,
	viewErrorHandler viewErrs.IErrorHandler,
) *Handler {
	return &Handler{
		PendingBookingsService: pendingBookingsService,
		ViewErrorHandler:       viewErrorHandler,
	}
}

func (h *Handler) Handle(c *gin.Context) {
	fmt.Printf("z requestu: %v", c.Request.PostForm)

	var form dto.PendingBookingForm
	if err := c.ShouldBind(&form); err != nil {
		viewErrs.ErrBadRequest(c, "bookings_pending_form.tmpl", err)
		return
	}

	parsedUUID, err := uuid.Parse(form.ClassID)
	if err != nil {
		viewErrs.ErrBadRequest(c, "bookings_pending_form.tmpl", err)
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
		h.ViewErrorHandler.Handle(c, "bookings_pending_form.tmpl", err)

		return
	}

	view := dto.PendingBookingView{
		ClassID: classID,
	}

	c.HTML(http.StatusOK, "pending_booking.tmpl", view)
}
