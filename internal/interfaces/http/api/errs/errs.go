package apiErrs

import "github.com/gin-gonic/gin"

type IErrorHandler interface {
	HandleJSONError(ctx *gin.Context, err error)
}
