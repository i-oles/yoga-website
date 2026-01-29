package dto

import (
	"fmt"
	"time"

	"main/internal/domain/models"
	"main/pkg/converter"

	"github.com/google/uuid"
)

type ActivatePassRequest struct {
	Email         string `binding:"required,min=3,max=40" json:"email"`
	UsedBookings  int    `binding:"min=0" json:"used_bookings"`
	TotalBookings int    `binding:"min=1" json:"total_bookings"`
}

type PassDTO struct {
	ID             int         `json:"id"`
	Email          string      `json:"email"`
	UsedBookingIDs []uuid.UUID `json:"used_booking_ids"`
	TotalBookings  int         `json:"total_bookings"`
	UpdatedAt      time.Time   `json:"updated_at"`
	CreatedAt      time.Time   `json:"created_at"`
}

func ToPassDTO(pass models.Pass) (PassDTO, error) {
	cratedAtWarsawTime, err := converter.ConvertToWarsawTime(pass.CreatedAt)
	if err != nil {
		return PassDTO{}, fmt.Errorf("error while converting createdAt to warsaw time: %w", err)
	}

	updatedAtWarsawTime, err := converter.ConvertToWarsawTime(pass.UpdatedAt)
	if err != nil {
		return PassDTO{}, fmt.Errorf("error while converting createdAt to warsaw time: %w", err)
	}

	return PassDTO{
		ID:             pass.ID,
		Email:          pass.Email,
		UsedBookingIDs: pass.UsedBookingIDs,
		TotalBookings:  pass.TotalBookings,
		UpdatedAt:      updatedAtWarsawTime,
		CreatedAt:      cratedAtWarsawTime,
	}, nil
}
