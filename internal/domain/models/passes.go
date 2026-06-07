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

type PassItem struct {
	Status         PassStatus
	ClassStartTime *time.Time
}

type PassStatus int

const (
	BlankPassStatus PassStatus = iota
	PastPassStatus
	FuturePassStatus
)

type PassUpdate struct {
	UsedBookingIDs []uuid.UUID
	TotalBookings  *int
}

type PassActivationParams struct {
	Email              string
	UsedBookingsCount  int
	TotalBookingsCount int
}
