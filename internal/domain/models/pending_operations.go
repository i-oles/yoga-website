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

type PendingOperation struct {
	ID             uuid.UUID `db:"id"`
	ClassID        uuid.UUID `db:"class_id"`
	Operation      Operation `db:"operation"`
	Email          string    `db:"email"`
	FirstName      string    `db:"first_name"`
	LastName       *string   `db:"last_name"`
	AuthToken      string    `db:"auth_token"`
	TokenExpiresAt time.Time `db:"token_expires_at"`
	CreatedAt      time.Time `db:"created_at"`
}

type CreateParams struct {
	ClassID   uuid.UUID `json:"class_id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
}

type CancelParams struct {
	ClassID   uuid.UUID `json:"class_id"`
	FirstName string    `json:"first_name"`
	Email     string    `json:"email"`
}

type ConfirmationMsgParams struct {
	RecipientEmail   string
	RecipientName    string
	ConfirmationLink string
}
