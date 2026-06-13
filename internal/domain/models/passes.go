package models

import (
	"time"
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
