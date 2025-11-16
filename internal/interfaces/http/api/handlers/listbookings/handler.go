package listbookings

import (
	"main/internal/domain/repositories"
	"main/internal/interfaces/http/api/dto"
	apiErrs "main/internal/interfaces/http/api/errs"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	bookingsRepo    repositories.IBookings
	apiErrorHandler apiErrs.IErrorHandler
}

func NewHandler(
	bookingsRepo repositories.IBookings,
	apiErrorHandler apiErrs.IErrorHandler,
) *Handler {
	return &Handler{
		bookingsRepo:    bookingsRepo,
		apiErrorHandler: apiErrorHandler,
	}
}

func (h *Handler) Handle(c *gin.Context) {
	ctx := c.Request.Context()

	allBookings, err := h.bookingsRepo.List(ctx)
	if err != nil {
		h.apiErrorHandler.Handle(c, err)

		return
	}

	bookingsListResponse, err := dto.ToBookingsListResponse(allBookings)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DTOResponse: " + err.Error()})

		return
	}

	c.JSON(http.StatusOK, bookingsListResponse)
}
