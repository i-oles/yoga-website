package home

import (
	"main/internal/domain/services"
	viewErrs "main/internal/interfaces/http/html/errs"
	sharedDTO "main/internal/interfaces/http/shared/dto"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	classesService   services.IClassesService
	viewErrorHandler viewErrs.IErrorHandler
	isVacation       bool
}

func NewHandler(
	classesService services.IClassesService,
	viewErrorHandler viewErrs.IErrorHandler,
	isVacation bool,
) *Handler {
	return &Handler{
		classesService:   classesService,
		viewErrorHandler: viewErrorHandler,
		isVacation:       isVacation,
	}
}

func (h *Handler) Handle(c *gin.Context) {
	ctx := c.Request.Context()

	classes, err := h.classesService.GetAllClasses(ctx)
	if err != nil {
		h.viewErrorHandler.Handle(c, "err.tmpl", err)

		return
	}

	classesView, err := sharedDTO.ToClassesListDTO(classes)
	if err != nil {
		viewErrs.ErrDTOConversion(c, "err.tmpl", err)

		return
	}

	c.HTML(http.StatusOK, "index.html", gin.H{
		"Classes": classesView, "IsVacation": h.isVacation,
	})
}
