package activatepass

import (
	"net/http"

	"main/internal/domain/models"
	"main/internal/domain/services"
	"main/internal/interfaces/http/api/dto"
	apiErrs "main/internal/interfaces/http/api/errs"

	"github.com/gin-gonic/gin"
)

type handler struct {
	passesService   services.IPassesService
	apiErrorHandler apiErrs.IErrorHandler
}

func NewHandler(
	passesService services.IPassesService,
	apiErrorHandler apiErrs.IErrorHandler,
) *handler {
	return &handler{
		passesService:   passesService,
		apiErrorHandler: apiErrorHandler,
	}
}

func (h *handler) Handle(ginCtx *gin.Context) {
	var dtoActivatePassRequest dto.ActivatePassRequest

	err := ginCtx.ShouldBindJSON(&dtoActivatePassRequest)
	if err != nil {
		ginCtx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	params := models.PassActivationParams{
		Email:         dtoActivatePassRequest.Email,
		UsedBookings:  dtoActivatePassRequest.UsedBookings,
		TotalBookings: dtoActivatePassRequest.TotalBookings,
	}

	ctx := ginCtx.Request.Context()

	pass, err := h.passesService.ActivatePass(ctx, params)
	if err != nil {
		h.apiErrorHandler.Handle(ginCtx, err)

		return
	}

	passResp, err := dto.ToPassDTO(pass)
	if err != nil {
		ginCtx.JSON(http.StatusInternalServerError, gin.H{"error": "DTOResponse: " + err.Error()})

		return
	}

	ginCtx.JSON(http.StatusOK, passResp)
}
