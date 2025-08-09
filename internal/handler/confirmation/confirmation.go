package confirmation

import (
	"context"
	"errors"
	"fmt"
	"main/internal/errs"
	"main/internal/repository"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

const (
	recordExistsCode = "23505"
)

type confirmation struct {
	//TODO: change names
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
	authToken := c.Query("auth_token")

	confirmationData, err := h.confirmBooking(ctx, authToken)
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

func (h *Handler) confirmBooking(ctx context.Context, authToken string) (confirmation, error) {
	bookingOpt, err := h.pendingBookingsRepo.Get(ctx, authToken)
	if err != nil {
		return confirmation{}, fmt.Errorf("error while getting pending booking: %w", err)
	}

	if !bookingOpt.Exists() {
		return confirmation{}, errors.New("invalid or expired confirmation link")
	}

	booking := bookingOpt.Get()

	fmt.Println(booking.ClassID.String())

	confirmedBooking := repository.ConfirmedBooking{
		ID:        uuid.New(),
		ClassID:   booking.ClassID,
		FirstName: booking.FirstName,
		LastName:  *booking.LastName,
		Email:     booking.Email,
		CreatedAt: time.Now(),
	}

	err = h.confirmedBookingsRepo.Insert(ctx, confirmedBooking)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == recordExistsCode {
			return confirmation{}, errs.ErrAlreadyBooked(booking.Email)
		}

		return confirmation{}, fmt.Errorf("error while inserting booking: %w", err)
	}

	err = h.pendingBookingsRepo.Delete(ctx, authToken)
	if err != nil {
		return confirmation{}, fmt.Errorf("error while deleting pending booking: %w", err)
	}

	err = h.classRepo.DecrementMaxCapacity(booking.ClassID)
	if err != nil {
		return confirmation{}, fmt.Errorf("error while decrementing class max capacity: %w", err)
	}

	class, err := h.classRepo.Get(booking.ClassID)
	if err != nil {
		return confirmation{}, fmt.Errorf("error while getting class: %w", err)
	}

	return confirmation{
		classType: class.ClassCategory,
		date:      fmt.Sprintf("%d %s %d", class.StartTime.Day(), class.StartTime.Month(), class.StartTime.Year()),
		hour:      fmt.Sprintf("%d:%02d", class.StartTime.Hour(), class.StartTime.Minute()),
		place:     class.Location,
	}, nil
}
