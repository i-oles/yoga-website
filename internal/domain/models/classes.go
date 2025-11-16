package models

import (
	"time"

	"github.com/google/uuid"
)

type Class struct {
	ID          uuid.UUID
	StartTime   time.Time
	ClassLevel  string
	ClassName   string
	MaxCapacity int
	Location    string
}

type ClassWithCurrentCapacity struct {
	ID              uuid.UUID
	StartTime       time.Time
	ClassLevel      string
	ClassName       string
	CurrentCapacity int
	MaxCapacity     int
	Location        string
}

type UpdateClass struct {
	StartTime   *time.Time
	ClassLevel  *string
	ClassName   *string
	MaxCapacity *int
	Location    *string
}
