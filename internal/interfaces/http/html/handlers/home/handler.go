package home

import (
	"net/http"

	"main/internal/domain/services"
	viewErrs "main/internal/interfaces/http/html/errs"
	sharedDTO "main/internal/interfaces/http/shared/dto"

	"github.com/gin-gonic/gin"
)

const classViewLimit = 4

type handler struct {
	classesService   services.IClassesService
	viewErrorHandler viewErrs.IErrorHandler
	isVacation       bool
}

func NewHandler(
	classesService services.IClassesService,
	viewErrorHandler viewErrs.IErrorHandler,
	isVacation bool,
) *handler {
	return &handler{
		classesService:   classesService,
		viewErrorHandler: viewErrorHandler,
		isVacation:       isVacation,
	}
}

func (h *handler) Handle(ginCtx *gin.Context) {
	ctx := ginCtx.Request.Context()

	limit := classViewLimit

	classes, err := h.classesService.ListClasses(ctx, true, &limit)
	if err != nil {
		h.viewErrorHandler.Handle(ginCtx, "err.tmpl", err)

		return
	}

	classesView, err := sharedDTO.ToClassesWithCurrentCapacityDTO(classes)
	if err != nil {
		viewErrs.ErrDTOConversion(ginCtx, "err.tmpl", err)

		return
	}

	ginCtx.HTML(http.StatusOK, "index.html", gin.H{
		"Classes":    classesView,
		"IsVacation": h.isVacation,
	})
}
