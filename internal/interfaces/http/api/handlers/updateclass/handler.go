package updateclass

import (
	"main/internal/domain/models"
	"main/internal/domain/services"
	"main/internal/interfaces/http/api/dto"
	apiErrs "main/internal/interfaces/http/api/errs"
	sharedDTO "main/internal/interfaces/http/shared/dto"
	"net/http"

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
	var dtoUpdateClass dto.UpdateClassRequest

	err := c.ShouldBindJSON(&dtoUpdateClass)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	var uri dto.UpdateClassURI

	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	parsedUUID, err := uuid.Parse(uri.ClassID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		
		return
	}

	ctx := c.Request.Context()

	update := models.UpdateClass{
		StartTime: dtoUpdateClass.StartTime,
		ClassLevel: dtoUpdateClass.ClassLevel,
		ClassName: dtoUpdateClass.ClassName,
		CurrentCapacity: dtoUpdateClass.CurrentCapacity,
		MaxCapacity: dtoUpdateClass.MaxCapacity,
		Location: dtoUpdateClass.Location,
	}

	updatedClass, err := h.classesService.UpdateClass(ctx, parsedUUID, update)
	if err != nil {
		h.apiErrorHandler.Handle(c, err)

		return
	}

	resp, err := sharedDTO.ToClassDTO(updatedClass)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DTOResponse: " + err.Error()})

		return
	}

	c.JSON(http.StatusOK, resp)
}
