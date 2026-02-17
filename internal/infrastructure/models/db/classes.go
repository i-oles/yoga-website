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

func SQLClassFromDomain(class models.Class) SQLClass {
	return SQLClass{
		ID:          class.ID,
		StartTime:   class.StartTime,
		ClassLevel:  class.ClassLevel,
		ClassName:   class.ClassName,
		MaxCapacity: class.MaxCapacity,
		Location:    class.Location,
	}
}
