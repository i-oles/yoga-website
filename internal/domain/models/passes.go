package models

import (
	"time"

	"github.com/google/uuid"
)

type Pass struct {
	ID             int
	Email          string
	UsedBookingIDs []uuid.UUID
	TotalBookings  int
	UpdatedAt      time.Time
	CreatedAt      time.Time
}

type PassUpdate struct {
	UsedBookingIDs []uuid.UUID
	TotalBookings  *int
}

type PassActivationParams struct {
	Email         string
	UsedBookings  int
	TotalBookings int
}
