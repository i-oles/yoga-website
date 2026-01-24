package models

import (
	"time"
)

type Pass struct {
	ID           int
	Email        string
	UsedCredits  int
	TotalCredits int
	CreatedAt    time.Time
}

type PassActivationParams struct {
	Email        string
	UsedCredits  int
	TotalCredits int
}
