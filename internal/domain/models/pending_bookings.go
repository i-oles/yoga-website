package models

import (
	"time"

	"github.com/google/uuid"
)

type Operation string

type PendingBooking struct {
	ID                uuid.UUID
	ClassID           uuid.UUID
	Email             string
	FirstName         string
	LastName          string
	ConfirmationToken string
	CreatedAt         time.Time
}

type PendingBookingParams struct {
	ClassID   uuid.UUID `json:"class_id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
}

type ConfirmationCreateMsg struct {
	RecipientEmail         string
	RecipientFirstName     string
	ConfirmationCreateLink string
}

type ConfirmationCancelMsg struct {
	RecipientEmail         string
	RecipientFirstName     string
	ConfirmationCancelLink string
}
