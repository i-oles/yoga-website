package models

import (
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
	ID              uuid.UUID  `db:"id"`
	DayOfWeek       string     `db:"day_of_week"`
	StartTime       time.Time  `db:"start_time"`
	ClassLevel      ClassLevel `db:"class_level"`
	ClassCategory   string     `db:"class_category"`
	CurrentCapacity int        `db:"current_capacity"`
	MaxCapacity     int        `db:"max_capacity"`
	Location        string     `db:"location"`
}
