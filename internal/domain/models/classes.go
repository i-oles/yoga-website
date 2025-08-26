package models

import (
	"time"

	"github.com/google/uuid"
)

type Class struct {
	ID              uuid.UUID `db:"id"`
	StartTime       time.Time `db:"start_time"`
	ClassLevel      string    `db:"class_level"`
	ClassName       string    `db:"class_name"`
	CurrentCapacity int       `db:"current_capacity"`
	MaxCapacity     int       `db:"max_capacity"`
	Location        string    `db:"location"`
}
