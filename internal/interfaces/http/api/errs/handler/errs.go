package handler

import (
	"errors"
	"net/http"

	domainErrs "main/internal/domain/errs/api"

	"github.com/gin-gonic/gin"
)

type ErrorHandler struct{}

func NewErrorHandler() ErrorHandler {
	return ErrorHandler{}
}

func (e ErrorHandler) Handle(c *gin.Context, err error) {
	var apiError *domainErrs.APIError
	if errors.As(err, &apiError) {
		switch apiError.Code {
		case domainErrs.BadRequestCode:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case domainErrs.ConflictCode:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case domainErrs.NotFoundCode:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		return
	}

	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

	return
}
