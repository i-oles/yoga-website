package home

import (
	"main/internal/domain/services"
	"main/internal/interfaces/http/api/dto"
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
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	classesResp, err := dto.models.ToClassesListResponse(classes)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	c.HTML(http.StatusOK, "classes.html", gin.H{
		"Classes": classesResp,
	})
}
