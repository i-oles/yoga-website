package db

import (
	"time"

	"main/internal/domain/models"

	"github.com/google/uuid"
)

type SQLClass struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	StartTime   time.Time `gorm:"not null"`
	ClassLevel  string    `gorm:"not null"`
	ClassName   string    `gorm:"not null"`
	MaxCapacity int       `gorm:"not null"`
	Location    string    `gorm:"not null"`
}

func (SQLClass) TableName() string {
	return "classes"
}

func (s SQLClass) ToDomain() models.Class {
	return models.Class{
		ID:          s.ID,
		StartTime:   s.StartTime,
		ClassLevel:  s.ClassLevel,
		ClassName:   s.ClassName,
		MaxCapacity: s.MaxCapacity,
		Location:    s.Location,
	}
}

func SQLClassFromDomain(c models.Class) SQLClass {
	return SQLClass{
		ID:          c.ID,
		StartTime:   c.StartTime,
		ClassLevel:  c.ClassLevel,
		ClassName:   c.ClassName,
		MaxCapacity: c.MaxCapacity,
		Location:    c.Location,
	}
}
