package models

import (
	"time"

	"github.com/google/uuid"
)

type Operation string

const (
	CreateBooking Operation = "create_booking"
	CancelBooking Operation = "cancel_booking"
)

type PendingBooking struct {
	ID        uuid.UUID `db:"id"`
	ClassID   uuid.UUID `db:"class_id"`
	Operation Operation `db:"operation"`
	Email     string    `db:"email"`
	FirstName string    `db:"first_name"`
	//TODO: remove pointer when will be no pending operation for cancel
	LastName          *string   `db:"last_name"`
	ConfirmationToken string    `db:"confirmation_token"`
	CreatedAt         time.Time `db:"created_at"`
}

type PendingBookingParams struct {
	ClassID   uuid.UUID `json:"class_id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
}

type CancelBookingParams struct {
	ClassID uuid.UUID `json:"class_id"`
	Email   string    `json:"email"`
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
