package cancel

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Handle(c *gin.Context) {
	//TODO: /pending endpoints do not need this - they should take id from url
	id := c.PostForm("id")

	c.HTML(http.StatusOK, "cancel.tmpl", gin.H{"ID": id})
}
