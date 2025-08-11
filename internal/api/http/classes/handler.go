package classes

import (
	"main/internal/domain/models"
	"main/internal/domain/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	classesService services.IServiceClasses
}

func NewHandler(classesService services.IServiceClasses) *Handler {
	return &Handler{classesService: classesService}
}

func (h *Handler) Handle(c *gin.Context) {
	ctx := c.Request.Context()

	classes, err := h.classesService.GetAllClasses(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	classesMapping := map[string][]models.Class{
		"Classes": classes,
	}

	c.HTML(http.StatusOK, "classes.html", classesMapping)
}
