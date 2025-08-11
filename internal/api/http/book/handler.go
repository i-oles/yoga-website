package book

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
	id := c.PostForm("id")

	c.HTML(http.StatusOK, "book.tmpl", gin.H{"ID": id})
}
