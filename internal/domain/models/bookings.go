package models

import (
	"time"

	"github.com/google/uuid"
)

type Booking struct {
	ID                uuid.UUID
	ClassID           uuid.UUID
	FirstName         string
	LastName          string
	Email             string
	CreatedAt         time.Time
	ConfirmationToken string
	RemindedAt        *time.Time
	Class             *Class
}
