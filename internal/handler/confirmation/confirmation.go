package confirmation

import (
	"errors"
	"fmt"
	"main/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

const (
	recordExistsCode = "23505"
)

type Handler struct {
	confirmedBookingsRepo repository.ConfirmedBookings
	pendingBookingsRepo   repository.PendingBookings
}

func NewHandler(
	confirmedBookingsRepo repository.ConfirmedBookings,
	pendingBookingsRepo repository.PendingBookings,
) *Handler {
	return &Handler{
		confirmedBookingsRepo: confirmedBookingsRepo,
		pendingBookingsRepo:   pendingBookingsRepo,
	}
}

func (h *Handler) Handle(c *gin.Context) {
	ctx := c.Request.Context()
	token := c.PostForm("token")

	bookingOpt, err := h.pendingBookingsRepo.Get(ctx, token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	if !bookingOpt.Exists() {
		c.JSON(http.StatusNotFound, gin.H{"error": "invalid or expired confirmation link"})

		return
	}

	booking := bookingOpt.Get()

	err = h.confirmedBookingsRepo.Insert(
		ctx,
		booking.ClassID,
		booking.Name,
		booking.LastName,
		booking.Email,
	)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			if pgErr.Code == recordExistsCode {
				c.HTML(http.StatusConflict, "book.tmpl", gin.H{
					"ID":    booking.ClassID,
					"Error": fmt.Errorf("'%s' already booked in this class", booking.Email),
				})

				return
			}
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

}
