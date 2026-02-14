package dto

import (
	"time"
)

type CreateClassRequest struct {
	StartTime   time.Time `binding:"required" json:"start_time"  time_format:"2006-01-02T15:04:05Z07:00"`
	ClassLevel  string    `binding:"required,min=3,max=40" json:"class_level"`
	ClassName   string    `binding:"required,min=3,max=40" json:"class_name"`
	MaxCapacity int       `binding:"gte=1" json:"max_capacity"`
	Location    string    `binding:"required" json:"location"`
}

type GetClassesRequest struct {
	OnlyUpcomingClasses bool `json:"only_upcoming_classes"`
	ClassesLimit        *int `json:"classes_limit"`
}

type DeleteClassRequest struct {
	Message *string `binding:"omitempty,min=1,max=250" json:"message"`
}

type UpdateClassRequest struct {
	StartTime   *time.Time `json:"start_time"`
	ClassLevel  *string    `json:"class_level"`
	ClassName   *string    `json:"class_name"`
	MaxCapacity *int       `json:"max_capacity"`
	Location    *string    `json:"location"`
}

type UpdateClassURI struct {
	ClassID string `binding:"required" uri:"class_id"`
}
