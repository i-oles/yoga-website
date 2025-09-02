package allconfirmedbookings

import (
	"main/internal/api/http/dto"
	"main/internal/api/http/err/handler"
	"main/internal/domain/repositories"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	confirmedBookingsRepo repositories.IConfirmedBookings
	errorHandler          handler.IErrorHandler
}

func NewHandler(
	confirmedBookingsRepo repositories.IConfirmedBookings,
	errorHandler handler.IErrorHandler,
) *Handler {
	return &Handler{
		confirmedBookingsRepo: confirmedBookingsRepo,
		errorHandler:          errorHandler,
	}
}

func (h *Handler) Handle(c *gin.Context) {
	ctx := c.Request.Context()

	allConfirmedBookings, err := h.confirmedBookingsRepo.GetAll(ctx)
	if err != nil {
		h.errorHandler.HandleJSONError(c, err)

		return
	}

	confirmedBookingsResponse, err := dto.ToConfirmedBookingListResponse(allConfirmedBookings)
	if err != nil {
		h.errorHandler.HandleJSONError(c, err)

		return
	}

	c.JSON(http.StatusOK, confirmedBookingsResponse)
}
