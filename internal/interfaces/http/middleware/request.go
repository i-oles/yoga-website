package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequestID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := uuid.NewString()

		ctx.Set("request_id", id)
		ctx.Writer.Header().Set("X-Request-ID", id)

		ctx.Next()
	}
}
