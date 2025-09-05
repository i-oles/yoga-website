package models

import (
	"time"

	"github.com/google/uuid"
)

type Class struct {
	ID              uuid.UUID `json:"id" db:"id"`
	StartTime       time.Time `json:"start_time" db:"start_time"`
	ClassLevel      string    `json:"class_level" db:"class_level"`
	ClassName       string    `json:"class_name" db:"class_name"`
	CurrentCapacity int       `json:"current_capacity" db:"current_capacity"`
	MaxCapacity     int       `json:"max_capacity" db:"max_capacity"`
	Location        string    `json:"location" db:"location"`
	Bookings        []Booking `json:"bookings" db:"bookings"`
}
