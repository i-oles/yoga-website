package addclasses

import (
	"main/internal/api/http/dto"
	"main/internal/domain/models"
	"main/internal/domain/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	classesService services.IClassesService
}

func NewHandler(classesService services.IClassesService) *Handler {
	return &Handler{classesService: classesService}
}

func (h *Handler) Handle(c *gin.Context) {
	ctx := c.Request.Context()

	var classes []models.Class

	err := c.ShouldBindJSON(&classes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
