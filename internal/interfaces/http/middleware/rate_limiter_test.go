package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func TestGlobalRateLimit(t *testing.T) {
	tests := []struct {
		name           string
		requests       int
		expectedStatus int
		delayBetween   time.Duration
	}{
		{
			name:           "Success: 1 request",
			requests:       1,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Failure: 3 requests",
			requests:       3,
			expectedStatus: http.StatusTooManyRequests,
			delayBetween:   0,
		},
		{
			name:           "Success: 2 reqests (burst)",
			requests:       2,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)

			limiter := rate.NewLimiter(rate.Limit(1), 2)

			handler := GlobalRateLimit(limiter)

			var lastStatus int

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			router := gin.New()
			router.Use(handler)

			for range tt.requests {
				router.GET("/test", func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "ok"})
				})

				req, _ := http.NewRequestWithContext(c, http.MethodGet, "/test", nil)
				router.ServeHTTP(w, req)

				lastStatus = w.Code
			}

			if lastStatus != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, lastStatus)
			}
		})
	}
}
