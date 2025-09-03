package pendingbookingform

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
	c.HTML(http.StatusOK, "pending_booking_form.tmpl", gin.H{"ID": c.PostForm("id")})
}
