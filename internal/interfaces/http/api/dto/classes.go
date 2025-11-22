package dto

import (
	"time"
)

type CreateClassRequest struct {
	StartTime   time.Time `binding:"required" time_format:"2006-01-02T15:04:05Z07:00" json:"start_time"`
	ClassLevel  string    `binding:"required,min=3,max=40" json:"class_level"`
	ClassName   string    `binding:"required,min=3,max=40" json:"class_name"`
	MaxCapacity int       `binding:"gte=1"					json:"max_capacity"`
	Location    string    `binding:"required" json:"max_capacity" `
}

type GetClassesRequest struct {
	OnlyUpcomingClasses bool `json:"only_upcoming_classes"`
	ClassesLimit        *int `json:"classes_limit"`
}

type DeleteClassRequest struct {
	ReasonMsg string `json:"reason_msg" binding:"required,min=1,max=250"`
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
