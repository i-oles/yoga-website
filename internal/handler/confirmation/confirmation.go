package confirmation

import (
	"context"
	"errors"
	"fmt"
	"main/internal/errs"
	"main/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

const (
	recordExistsCode = "23505"
)

type confirmation struct {
	classType string
	date      string
	hour      string
	place     string
}

type Handler struct {
	confirmedBookingsRepo repository.ConfirmedBookings
	pendingBookingsRepo   repository.PendingBookings
	classRepo             repository.Classes
	errorHandler          errs.ErrorHandler
}

func NewHandler(
	confirmedBookingsRepo repository.ConfirmedBookings,
	pendingBookingsRepo repository.PendingBookings,
	classRepo repository.Classes,
	errorHandler errs.ErrorHandler,
) *Handler {
	return &Handler{
		confirmedBookingsRepo: confirmedBookingsRepo,
		pendingBookingsRepo:   pendingBookingsRepo,
		classRepo:             classRepo,
		errorHandler:          errorHandler,
	}
}

func (h *Handler) Handle(c *gin.Context) {
	ctx := c.Request.Context()
	token := c.Query("token")

	confirmationData, err := h.confirmBooking(ctx, token)
	if err != nil {
		h.errorHandler.Handle(c, "err.tmpl", err)

		return
	}

	c.HTML(http.StatusOK, "confirmation.tmpl", gin.H{
		"classType": confirmationData.classType,
		"place":     confirmationData.place,
		"date":      confirmationData.date,
		"hour":      confirmationData.hour,
	})
}

func (h *Handler) confirmBooking(ctx context.Context, token string) (confirmation, error) {
	bookingOpt, err := h.pendingBookingsRepo.Get(ctx, token)
	if err != nil {
		return confirmation{}, fmt.Errorf("error while getting pending booking: %w", err)
	}

	if !bookingOpt.Exists() {
		return confirmation{}, errors.New("invalid or expired confirmation link")
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
		if errors.As(err, &pgErr) && pgErr.Code == recordExistsCode {
			return confirmation{}, errs.ErrAlreadyBooked(booking.Email)
		}

		return confirmation{}, fmt.Errorf("error while inserting booking: %w", err)
	}

	err = h.pendingBookingsRepo.Delete(ctx, token)
	if err != nil {
		return confirmation{}, fmt.Errorf("error while deleting pending booking: %w", err)
	}

	err = h.classRepo.DecrementSpotsLeft(booking.ClassID)
	if err != nil {
		return confirmation{}, fmt.Errorf("error while decrementing spots left: %w", err)
	}

	return confirmation{
		classType: booking.ClassType,
		date:      fmt.Sprintf("%d %s %d", booking.Date.Day(), booking.Date.Month(), booking.Date.Year()),
		hour:      fmt.Sprintf("%d:%02d", booking.Date.Hour(), booking.Date.Minute()),
		place:     booking.Place,
	}, nil
}
