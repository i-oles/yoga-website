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

//TODO implement update class, and add swagger to easy update
//type UpdateClass struct {
//	ID        int                          `db:"id"`
//	Datetime  optional.Optional[time.Time] `db:"datetime"`
//	Day       optional.Optional[string]    `db:"day"`
//	Level     optional.Optional[string]    `db:"level"`
//	Type      optional.Optional[string]    `db:"type"`
//	SpotsLeft optional.Optional[int]       `db:"spotsLeft"`
//	Place     optional.Optional[string]    `db:"place"`
//}

// TODO: here should be ctx added
type Classes interface {
	GetAll() ([]Class, error)
	Get(id int) (Class, error)
	DecrementSpotsLeft(id int) error
}

type ConfirmedBookings interface {
	Insert(ctx context.Context, classID int, name, lastName, email string) error
}

type PendingBooking struct {
	ClassID   int       `db:"class_id"`
	ClassType string    `db:"class_type"`
	Place     string    `db:"place"`
	Date      time.Time `db:"date"`
	Name      string    `db:"name"`
	LastName  string    `db:"last_name"`
	Email     string    `db:"email"`
	Token     string    `db:"token"`
	ExpiresAt time.Time `db:"expires_at"`
}

type PendingBookings interface {
	Insert(ctx context.Context, booking PendingBooking) error
	Get(ctx context.Context, token string) (optional.Optional[PendingBooking], error)
	Delete(ctx context.Context, token string) error
}

type Booking struct {
	ID        int       `db:"id"`
	ClassID   int       `db:"class_id"`
	Name      string    `db:"name"`
	LastName  string    `db:"last_name"`
	Email     string    `db:"email"`
	CreatedAt time.Time `db:"created_at"`
}
