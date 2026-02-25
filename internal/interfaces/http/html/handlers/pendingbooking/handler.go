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

func (h *handler) Handle(ginCtx *gin.Context) {
	var form dto.PendingBookingForm
	if err := ginCtx.ShouldBind(&form); err != nil {
		viewErrs.ErrBadRequest(ginCtx, "pending_booking_form.tmpl", err)

		return
	}

	classID, err := uuid.Parse(form.ClassID)
	if err != nil {
		viewErrs.ErrBadRequest(ginCtx, "pending_booking_form.tmpl", err)

		return
	}

	pendingBookingParams := models.PendingBookingParams{
		ClassID:   classID,
		FirstName: form.FirstName,
		LastName:  form.LastName,
		Email:     strings.ToLower(form.Email),
	}

	ctx := ginCtx.Request.Context()

	err = h.PendingBookingsService.CreatePendingBooking(ctx, pendingBookingParams)
	if err != nil {
		h.ViewErrorHandler.Handle(ginCtx, "pending_booking_form.tmpl", err)

		return
	}

	ginCtx.HTML(http.StatusOK, "pending_booking.tmpl", gin.H{"ClassID": classID})
}
