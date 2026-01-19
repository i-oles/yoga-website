package models

import (
	"time"
)

type Pass struct {
	ID           int
	Email        string
	Credits      int
	TotalCredits int
	CreatedAt    time.Time
}
