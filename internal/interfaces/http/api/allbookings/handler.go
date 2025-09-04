package allbookings

import (
	"main/internal/domain/repositories"
	"main/internal/interfaces/http/api/dto"
	"main/internal/interfaces/http/err/handler"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	bookingsRepo repositories.IBookings
	errorHandler handler.IErrorHandler
}

func NewHandler(
	bookingsRepo repositories.IBookings,
	errorHandler handler.IErrorHandler,
) *Handler {
	return &Handler{
		bookingsRepo: bookingsRepo,
		errorHandler: errorHandler,
	}
}

func (h *Handler) Handle(c *gin.Context) {
	ctx := c.Request.Context()

	allBookings, err := h.bookingsRepo.GetAll(ctx)
	if err != nil {
		h.errorHandler.HandleJSONError(c, err)

		return
	}

	bookingsListResponse, err := dto.ToBookingsListResponse(allBookings)
	if err != nil {
		h.errorHandler.HandleJSONError(c, err)

		return
	}

	c.JSON(http.StatusOK, bookingsListResponse)
}
