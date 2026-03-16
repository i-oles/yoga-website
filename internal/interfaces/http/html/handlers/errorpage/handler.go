package errorpage

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Handle(c *gin.Context) {
	c.HTML(http.StatusInternalServerError, "err.tmpl", gin.H{
		"Error": "Coś poszło nie tak... Skontaktuj się ze mną. ID błędu: " + c.GetHeader("request_id"),
	})
}
