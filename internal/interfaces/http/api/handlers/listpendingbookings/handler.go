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

func (h *handler) Handle(ginCtx *gin.Context) {
	ctx := ginCtx.Request.Context()

	allPendingBookings, err := h.pendingBookingsRepo.List(ctx)
	if err != nil {
		h.apiErrorHandler.Handle(ginCtx, err)

		return
	}

	pendingBookingsListResponse, err := dto.ToPendingBookingsListResponse(allPendingBookings)
	if err != nil {
		ginCtx.JSON(http.StatusInternalServerError, gin.H{"error": "DTOResponse: " + err.Error()})

		return
	}

	ginCtx.JSON(http.StatusOK, pendingBookingsListResponse)
}
