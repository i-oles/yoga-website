package classes

import (
	"main/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	classesRepo repository.Classes
}

func NewHandler(classesRepo repository.Classes) *Handler {
	return &Handler{classesRepo: classesRepo}
}

func (h *Handler) Handle(c *gin.Context) {
	classes, err := h.classesRepo.GetCurrentMonthClasses()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	classesMapping := map[string][]repository.Class{
		"Classes": classes,
	}

	c.HTML(http.StatusOK, "classes.html", classesMapping)
}
