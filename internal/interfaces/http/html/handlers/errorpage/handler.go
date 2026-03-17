package errorpage

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Handle(c *gin.Context) {
	err := fmt.Errorf("error_id: %s", c.GetString("request_id"))
	c.HTML(http.StatusInternalServerError, "err.tmpl", gin.H{
		"Error": err.Error(),
	})
}
