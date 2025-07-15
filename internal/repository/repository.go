package repository

import (
	"context"
	"main/pkg/optional"
	"time"
)

type Class struct {
	ID        int       `db:"id"`
	Datetime  time.Time `db:"datetime"`
	Day       string    `db:"day"`
	Level     string    `db:"level"`
	Type      string    `db:"type"`
	SpotsLeft int       `db:"spotsLeft"`
	Place     string    `db:"place"`
}

type UpdateClass struct {
	ID        int       `db:"id"`
	Datetime  time.Time `db:"datetime"`
	Day       string    `db:"day"`
	Level     string    `db:"level"`
	Type      string    `db:"type"`
	SpotsLeft int       `db:"spotsLeft"`
	Place     string    `db:"place"`
}

type Classes interface {
	GetAll() ([]Class, error)
	Get(id int) (Class, error)
}

type ConfirmedBookings interface {
	Insert(ctx context.Context, classID int, name, lastName, email string) error
}

type PendingBooking struct {
	ClassID   int       `db:"class_id"`
	Name      string    `db:"name"`
	LastName  string    `db:"last_name"`
	Email     string    `db:"email"`
	Token     string    `db:"token"`
	ExpiresAt time.Time `db:"expires_at"`
}

type PendingBookings interface {
	Insert(ctx context.Context, booking PendingBooking) error
	Get(ctx context.Context, token string) (optional.Optional[PendingBooking], error)
}

type Booking struct {
	ID        int       `db:"id"`
	ClassID   int       `db:"class_id"`
	Name      string    `db:"name"`
	LastName  string    `db:"last_name"`
	Email     string    `db:"email"`
	CreatedAt time.Time `db:"created_at"`
}
