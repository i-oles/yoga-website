package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func GlobalRateLimit(l *rate.Limiter) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !l.Allow() {
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Too Many Requests",
			})

			return
		}

		ctx.Next()
	}
}
