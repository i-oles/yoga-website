package models

import (
	"time"

	"github.com/google/uuid"
)

type Pass struct {
	ID             int
	Email          string
	UsedBookingIDs []uuid.UUID
	TotalCredits   int
	UpdatedAt      time.Time
	CreatedAt      time.Time
}

type PassUpdate struct {
	UsedBookingIDs []uuid.UUID
	TotalCredits   *int
}

type PassActivationParams struct {
	Email        string
	UsedCredits  int
	TotalCredits int
}
