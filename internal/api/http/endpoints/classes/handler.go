package classes

import (
	"fmt"
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

	classes, err := h.classesService.GetAllClasses(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	classesResp, err := toClassesListResponse(classes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.HTML(http.StatusOK, "classes.html", gin.H{
		"Classes": classesResp,
	})
}

func toClassesListResponse(classes []models.Class) ([]dto.ClassResponse, error) {
	classesResponse := make([]dto.ClassResponse, len(classes))
	for i, class := range classes {
		classResponse, err := dto.ToClassResponse(class)
		if err != nil {
			return nil, fmt.Errorf("could not convert class to classResponse: %w", err)
		}

		classesResponse[i] = classResponse
	}

	return classesResponse, nil
}
