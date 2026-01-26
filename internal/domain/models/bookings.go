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
	Class             *Class
}

type ConfirmationMsg struct {
	RecipientEmail     string
	RecipientFirstName string
	RecipientLastName  string
	ClassName          string
	ClassLevel         string
	StartTime          time.Time
	Location           string
	CancellationLink   string
	UsedPassCredits    int
	TotalPassCredits   int
}

type CancellationMsg struct {
	RecipientFirstName string
	RecipientEmail     string
	ClassName          string
	ClassLevel         string
	StartTime          time.Time
	Location           string
	UsedPassCredits    *int
	TotalPassCredits   *int
}
