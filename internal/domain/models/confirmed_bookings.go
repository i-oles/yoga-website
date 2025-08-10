package models

import (
	"time"

	"github.com/google/uuid"
)

type ConfirmedBooking struct {
	ID        uuid.UUID `db:"id"`
	ClassID   uuid.UUID `db:"class_id"`
	FirstName string    `db:"first_name"`
	LastName  string    `db:"last_name"`
	Email     string    `db:"email"`
	CreatedAt time.Time `db:"created_at"`
}
