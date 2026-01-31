package activatepass

import (
	"net/http"

	"main/internal/domain/models"
	"main/internal/domain/services"
	"main/internal/interfaces/http/api/dto"
	apiErrs "main/internal/interfaces/http/api/errs"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	passesService   services.IPassesService
	apiErrorHandler apiErrs.IErrorHandler
}

func NewHandler(
	passesService services.IPassesService,
	apiErrorHandler apiErrs.IErrorHandler,
) *Handler {
	return &Handler{
		passesService:   passesService,
		apiErrorHandler: apiErrorHandler,
	}
}

func (h *Handler) Handle(c *gin.Context) {
	var dtoActivatePassRequest dto.ActivatePassRequest

	err := c.ShouldBindJSON(&dtoActivatePassRequest)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	params := models.PassActivationParams{
		Email:         dtoActivatePassRequest.Email,
		UsedBookings:  dtoActivatePassRequest.UsedBookings,
		TotalBookings: dtoActivatePassRequest.TotalBookings,
	}

	ctx := c.Request.Context()

	pass, err := h.passesService.ActivatePass(ctx, params)
	if err != nil {
		h.apiErrorHandler.Handle(c, err)

		return
	}

	passResp, err := dto.ToPassDTO(pass)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DTOResponse: " + err.Error()})

		return
	}

	c.JSON(http.StatusOK, passResp)
}
