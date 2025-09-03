package createclasses

import (
	"main/internal/domain/models"
	"main/internal/domain/services"
	"main/internal/interfaces/http/api/dto"
	"main/internal/interfaces/http/err/handler"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	classesService services.IClassesService
	errorHandler   handler.IErrorHandler
}

func NewHandler(
	classesService services.IClassesService,
	errorHandler handler.IErrorHandler,
) *Handler {
	return &Handler{
		classesService: classesService,
		errorHandler:   errorHandler,
	}
}

func (h *Handler) Handle(c *gin.Context) {
	var dtoClasses []dto.ClassRequest

	err := c.ShouldBindJSON(&dtoClasses)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	classes := make([]models.Class, 0, len(dtoClasses))

	for _, dtoClass := range dtoClasses {
		class := models.Class{
			ID:              uuid.New(),
			StartTime:       dtoClass.StartTime.UTC(),
			ClassLevel:      dtoClass.ClassLevel,
			ClassName:       dtoClass.ClassName,
			CurrentCapacity: dtoClass.CurrentCapacity,
			MaxCapacity:     dtoClass.MaxCapacity,
			Location:        dtoClass.Location,
		}

		classes = append(classes, class)
	}

	ctx := c.Request.Context()

	createdClasses, err := h.classesService.CreateClasses(ctx, classes)
	if err != nil {
		h.errorHandler.HandleJSONError(c, err)

		return
	}

	classesResp, err := dto.ToClassesListResponse(createdClasses)
	if err != nil {
		h.errorHandler.HandleJSONError(c, err)

		return
	}

	c.JSON(http.StatusCreated, classesResp)
}
