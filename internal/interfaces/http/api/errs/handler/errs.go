package handler

import (
	"errors"
	domainErrs "main/internal/domain/errs"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorHandler struct{}

func NewErrorHandler() ErrorHandler {
	return ErrorHandler{}
}

func (e ErrorHandler) HandleJSONError(c *gin.Context, err error) {
	var classError *domainErrs.ClassError
	if errors.As(err, &classError) {
		switch classError.Code {
		case domainErrs.BadRequestCode:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		return
	}

	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

	return
}
