package dto

import (
	"time"
)

type CreateClassRequest struct {
	StartTime       time.Time `json:"start_time" binding:"required" time_format:"2006-01-02T15:04:05Z07:00"`
	ClassLevel      string    `json:"class_level" binding:"required,min=3,max=40"`
	ClassName       string    `json:"class_name" binding:"required,min=3,max=40"`
	CurrentCapacity int       `json:"current_capacity" binding:"gte=0"`
	MaxCapacity     int       `json:"max_capacity" binding:"gte=1"`
	Location        string    `json:"location" binding:"required"`
}

type GetClassesRequest struct {
	OnlyUpcomingClasses bool `json:"only_upcoming_classes"`
	ClassesLimit        *int `json:"classes_limit"`
}
