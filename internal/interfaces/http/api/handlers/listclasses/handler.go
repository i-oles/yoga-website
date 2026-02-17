package listclasses

import (
	"net/http"

	"main/internal/domain/services"
	"main/internal/interfaces/http/api/dto"
	apiErrs "main/internal/interfaces/http/api/errs"
	sharedDTO "main/internal/interfaces/http/shared/dto"

	"github.com/gin-gonic/gin"
)

type handler struct {
	classesService  services.IClassesService
	apiErrorHandler apiErrs.IErrorHandler
}

func NewHandler(
	classesService services.IClassesService,
	apiErrorHandler apiErrs.IErrorHandler,
) *handler {
	return &handler{
		classesService:  classesService,
		apiErrorHandler: apiErrorHandler,
	}
}

func (h *handler) Handle(ginCtx *gin.Context) {
	var dtoGetClasses dto.GetClassesRequest

	err := ginCtx.ShouldBindJSON(&dtoGetClasses)
	if err != nil {
		ginCtx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	ctx := ginCtx.Request.Context()

	classes, err := h.classesService.ListClasses(
		ctx,
		dtoGetClasses.OnlyUpcomingClasses,
		dtoGetClasses.ClassesLimit,
	)
	if err != nil {
		h.apiErrorHandler.Handle(ginCtx, err)

		return
	}

	classesResp, err := sharedDTO.ToClassesWithCurrentCapacityDTO(classes)
	if err != nil {
		ginCtx.JSON(http.StatusInternalServerError, gin.H{"error": "DTOResponse: " + err.Error()})

		return
	}

	ginCtx.JSON(http.StatusOK, classesResp)
}
