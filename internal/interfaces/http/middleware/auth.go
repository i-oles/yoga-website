package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Auth(secret string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		expected := "Bearer " + secret

		if authHeader != expected {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})

			return
		}

		ctx.Next()
	}
}
