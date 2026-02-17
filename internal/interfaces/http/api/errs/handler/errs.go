package handler

import (
	"errors"
	"net/http"

	domainErrs "main/internal/domain/errs/api"

	"github.com/gin-gonic/gin"
)

type errorHandler struct{}

func NewErrorHandler() errorHandler {
	return errorHandler{}
}

func (e errorHandler) Handle(ctx *gin.Context, err error) {
	var apiError *domainErrs.APIError
	if errors.As(err, &apiError) {
		switch apiError.Code {
		case domainErrs.BadRequestCode:
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case domainErrs.ConflictCode:
			ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case domainErrs.NotFoundCode:
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		return
	}

	ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}
