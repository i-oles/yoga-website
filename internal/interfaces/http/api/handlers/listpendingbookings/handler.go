package listpendingbookings

import (
	"net/http"

	"main/internal/domain/repositories"
	"main/internal/interfaces/http/api/dto"
	apiErrs "main/internal/interfaces/http/api/errs"

	"github.com/gin-gonic/gin"
)

type handler struct {
	pendingBookingsRepo repositories.IPendingBookings
	apiErrorHandler     apiErrs.IErrorHandler
}

func NewHandler(
	pendingBookingsRepo repositories.IPendingBookings,
	apiErrorHandler apiErrs.IErrorHandler,
) *handler {
	return &handler{
		pendingBookingsRepo: pendingBookingsRepo,
		apiErrorHandler:     apiErrorHandler,
	}
}

func (h *handler) Handle(c *gin.Context) {
	ctx := c.Request.Context()

	allPendingBookings, err := h.pendingBookingsRepo.List(ctx)
	if err != nil {
		h.apiErrorHandler.Handle(c, err)

		return
	}

	pendingBookingsListResponse, err := dto.ToPendingBookingsListResponse(allPendingBookings)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DTOResponse: " + err.Error()})

		return
	}

	c.JSON(http.StatusOK, pendingBookingsListResponse)
}
