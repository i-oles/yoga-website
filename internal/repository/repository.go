package repository

import (
	"context"
	"main/pkg/optional"
	"time"

	"github.com/google/uuid"
)

type ClassLevel string

const (
	beginner     ClassLevel = "beginner"
	intermediate ClassLevel = "intermediate"
	advanced     ClassLevel = "advanced"
)

type Class struct {
	ID            uuid.UUID  `db:"id"`
	DayOfWeek     string     `db:"day_of_week"`
	StartTime     time.Time  `db:"start_time"`
	ClassLevel    ClassLevel `db:"class_level"`
	ClassCategory string     `db:"class_category"`
	MaxCapacity   int        `db:"max_capacity"`
	Location      string     `db:"location"`
}

// TODO: here should be ctx added
type Classes interface {
	GetAll() ([]Class, error)
	Get(id uuid.UUID) (Class, error)
	DecrementMaxCapacity(id uuid.UUID) error
}

type ConfirmedBooking struct {
	ID        uuid.UUID `db:"id"`
	ClassID   uuid.UUID `db:"class_id"`
	FirstName string    `db:"first_name"`
	LastName  string    `db:"last_name"`
	Email     string    `db:"email"`
	CreatedAt time.Time `db:"created_at"`
}

type ConfirmedBookings interface {
	Insert(ctx context.Context, confirmedBooking ConfirmedBooking) error
}

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

type PendingBookings interface {
	Insert(ctx context.Context, booking PendingOperation) error
	Get(ctx context.Context, token string) (optional.Optional[PendingOperation], error)
	Delete(ctx context.Context, token string) error
}
