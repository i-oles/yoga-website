package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func RequestID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := strconv.FormatInt(time.Now().Unix(), 10)

		ctx.Set("request_id", id)
		ctx.Writer.Header().Set("X-Request-ID", id)

		ctx.Next()
	}
}
