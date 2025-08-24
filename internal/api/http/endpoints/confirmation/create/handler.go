package create

import (
	"main/internal/api/http/dto"
	"main/internal/api/http/err/handler"
	"main/internal/domain/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	confirmationService services.IConfirmationService
	errorHandler        handler.IErrorHandler
}

func NewHandler(
	confirmationService services.IConfirmationService,
	errorHandler handler.IErrorHandler,
) *Handler {
	return &Handler{
		confirmationService: confirmationService,
		errorHandler:        errorHandler,
	}
}

func (h *Handler) Handle(c *gin.Context) {
	var req dto.ConfirmationCreateRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	ctx := c.Request.Context()

	class, err := h.confirmationService.CreateBooking(ctx, req.Token)
	if err != nil {
		h.errorHandler.Handle(c, "err.tmpl", err)

		return
	}

	resp, err := dto.ToConfirmationCreateResponse(class)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.HTML(http.StatusOK, "confirmation_create.tmpl", resp)
}
