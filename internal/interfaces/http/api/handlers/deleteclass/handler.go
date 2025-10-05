package deleteclass

import (
	"main/internal/domain/services"
	"main/internal/interfaces/http/api/errs"
	"net/http"

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
	classIDStr := c.Param("class_id")

	classID, err := uuid.Parse(classIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	ctx := c.Request.Context()

	err = h.classesService.DeleteClass(ctx, classID)
	if err != nil {
		h.apiErrorHandler.Handle(c, err)

		return
	}

	c.JSON(http.StatusOK, gin.H{"class_id" : classID})
}
