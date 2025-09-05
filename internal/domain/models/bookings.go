package models

import (
	"time"

	"github.com/google/uuid"
)

type Booking struct {
	ID                uuid.UUID `db:"id"`
	ClassID           uuid.UUID `db:"class_id"`
	FirstName         string    `db:"first_name"`
	LastName          string    `db:"last_name"`
	Email             string    `db:"email"`
	CreatedAt         time.Time `db:"created_at"`
	ConfirmationToken string    `db:"confirmation_token"`
	Class             Class     `db:"class"`
}

type ConfirmationMsg struct {
	RecipientEmail     string
	RecipientFirstName string
	RecipientLastName  string
	ClassName          string
	ClassLevel         string
	WeekDay            string
	Hour               string
	Date               string
	Location           string
	CancellationLink   string
}

type ConfirmationToOwnerMsg struct {
	RecipientFirstName string
	RecipientLastName  string
	WeekDay            string
	Hour               string
	Date               string
}
