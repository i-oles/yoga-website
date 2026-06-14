package models

import (
	"time"

	"github.com/google/uuid"
)

type Pass struct {
	ID         int
	Email      string
	TotalSlots int
	UpdatedAt  time.Time
	CreatedAt  time.Time
}

type PassSlot struct {
	Status         PassStatus
	ClassStartTime *time.Time
}

type PassStatus int

const (
	BlankPassStatus PassStatus = iota
	PastPassStatus
	FuturePassStatus
)

type PassActivationParams struct {
	Email      string
	UsedSlots  int
	TotalSlots int
}

type PassActivation struct {
	Pass            Pass
	BookingIDsAdded []uuid.UUID
}
