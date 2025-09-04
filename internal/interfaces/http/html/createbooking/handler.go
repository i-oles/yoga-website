package createbooking

import (
	"main/internal/domain/services"
	"main/internal/interfaces/http/err/handler"
	"main/internal/interfaces/http/html/dto"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	bookingsService services.IBookingsService
	errorHandler    handler.IErrorHandler
}

func NewHandler(
	confirmationService services.IBookingsService,
	errorHandler handler.IErrorHandler,
) *Handler {
	return &Handler{
		bookingsService: confirmationService,
		errorHandler:    errorHandler,
	}
}

func (h *Handler) Handle(c *gin.Context) {
	var form dto.BookingCreateForm
	if err := c.ShouldBindQuery(&form); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	ctx := c.Request.Context()

	class, err := h.bookingsService.CreateBooking(ctx, form.Token)
	if err != nil {
		h.errorHandler.HandleHTMLError(c, "err.tmpl", err)

		return
	}

	view, err := dto.ToBookingView(class)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	c.HTML(http.StatusOK, "confirmation_create_booking.tmpl", view)
}
