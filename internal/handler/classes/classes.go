package classes

import (
	"html/template"
	"main/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ClassHandler struct {
	classesRepo repository.Classes
}

func NewHandler(classesRepo repository.Classes) *ClassHandler {
	return &ClassHandler{classesRepo: classesRepo}
}

func (h *ClassHandler) Handle(c *gin.Context) {
	classes, err := h.classesRepo.GetCurrentMonthClasses()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	classesMapping := map[string][]repository.Class{
		"Classes": classes,
	}

	tmpl, err := template.ParseFiles("templates/classes.html")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	err = tmpl.Execute(c.Writer, classesMapping)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}
}
