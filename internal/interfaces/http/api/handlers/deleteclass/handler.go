package deleteclass

import (
	"net/http"

	"main/internal/domain/services"
	"main/internal/interfaces/http/api/dto"
	apiErrs "main/internal/interfaces/http/api/errs"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	classesService  services.IClassesService
	apiErrorHandler apiErrs.IErrorHandler
}

func NewHandler(
	classesService services.IClassesService,
	apiErrorHandler apiErrs.IErrorHandler,
) *Handler {
	return &Handler{
		classesService:  classesService,
		apiErrorHandler: apiErrorHandler,
	}
}

func (h *Handler) Handle(c *gin.Context) {
	var dtoDeleteClassRequest dto.DeleteClassRequest

	err := c.ShouldBindJSON(&dtoDeleteClassRequest)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	classIDStr := c.Param("class_id")

	classID, err := uuid.Parse(classIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	ctx := c.Request.Context()

	err = h.classesService.DeleteClass(ctx, classID, dtoDeleteClassRequest.Message)
	if err != nil {
		h.apiErrorHandler.Handle(c, err)

		return
	}

	c.JSON(http.StatusOK, gin.H{"class_id": classID})
}
