package pendingbookingform

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type handler struct{}

func NewHandler() *handler {
	return &handler{}
}

func (h *handler) Handle(c *gin.Context) {
	c.HTML(http.StatusOK, "pending_booking_form.tmpl", gin.H{"ID": c.Param("class_id")})
}
