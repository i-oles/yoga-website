package classes

import (
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

	}

	c.HTML(http.StatusOK, "classes.html", classes)
}
