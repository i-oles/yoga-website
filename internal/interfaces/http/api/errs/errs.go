package apierrs

import "github.com/gin-gonic/gin"

type IErrorHandler interface {
	Handle(ctx *gin.Context, err error)
}
