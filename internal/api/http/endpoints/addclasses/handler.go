package addclasses

import (
	"main/internal/api/http/dto"
	"main/internal/domain/models"
	"main/internal/domain/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	classesService services.IClassesService
}

func NewHandler(classesService services.IClassesService) *Handler {
	return &Handler{classesService: classesService}
}

func (h *Handler) Handle(c *gin.Context) {
	ctx := c.Request.Context()

	var dtoClasses []dto.ClassRequest

	err := c.ShouldBindJSON(&dtoClasses)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	classes := make([]models.Class, 0, len(dtoClasses))

	for _, dtoClass := range dtoClasses {
		class := models.Class{
			ID:              uuid.New(),
			StartTime:       dtoClass.StartTime,
			ClassLevel:      dtoClass.ClassLevel,
			ClassName:       dtoClass.ClassName,
			CurrentCapacity: dtoClass.CurrentCapacity,
			MaxCapacity:     dtoClass.MaxCapacity,
			Location:        dtoClass.Location,
		}

		classes = append(classes, class)
	}

	createdClasses, err := h.classesService.CreateClasses(ctx, classes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	classesResp, err := dto.ToClassesListResponse(createdClasses)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusCreated, classesResp)
}
