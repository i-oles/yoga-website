package createclasses

import (
	"net/http"

	"main/internal/domain/models"
	"main/internal/domain/services"
	"main/internal/interfaces/http/api/dto"
	apiErrs "main/internal/interfaces/http/api/errs"
	sharedDTO "main/internal/interfaces/http/shared/dto"

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
	var dtoClasses []dto.CreateClassRequest

	err := c.ShouldBindJSON(&dtoClasses)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	classes := make([]models.Class, 0, len(dtoClasses))

	for _, dtoClass := range dtoClasses {
		class := models.Class{
			ID:          uuid.New(),
			StartTime:   dtoClass.StartTime.UTC(),
			ClassLevel:  dtoClass.ClassLevel,
			ClassName:   dtoClass.ClassName,
			MaxCapacity: dtoClass.MaxCapacity,
			Location:    dtoClass.Location,
		}

		classes = append(classes, class)
	}

	ctx := c.Request.Context()

	createdClasses, err := h.classesService.CreateClasses(ctx, classes)
	if err != nil {
		h.apiErrorHandler.Handle(c, err)

		return
	}

	classesResp, err := sharedDTO.ToClassesDTO(createdClasses)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DTOResponse: " + err.Error()})

		return
	}

	c.JSON(http.StatusCreated, classesResp)
}
