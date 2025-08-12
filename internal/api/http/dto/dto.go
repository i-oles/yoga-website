package dto

import (
	"fmt"
	"main/internal/domain/models"
	"time"

	"github.com/google/uuid"
)

type ClassLevel string

const (
	beginner     ClassLevel = "beginner"
	intermediate ClassLevel = "intermediate"
	advanced     ClassLevel = "advanced"
)

type ClassResponse struct {
	ID              uuid.UUID  `json:"id"`
	DayOfWeek       string     `json:"day_of_week"`
	StartTime       time.Time  `json:"start_time"`
	ClassLevel      ClassLevel `json:"class_level"`
	ClassCategory   string     `json:"class_category"`
	CurrentCapacity int        `json:"current_capacity"`
	MaxCapacity     int        `json:"max_capacity"`
	Location        string     `json:"location"`
}

func ToClassResponse(class models.Class) ClassResponse {
	return ClassResponse{
		ID:              class.ID,
		DayOfWeek:       class.DayOfWeek,
		StartTime:       class.StartTime,
		ClassLevel:      ClassLevel(class.ClassLevel),
		ClassCategory:   class.ClassCategory,
		CurrentCapacity: class.CurrentCapacity,
		MaxCapacity:     class.MaxCapacity,
		Location:        class.Location,
	}
}

type ConfirmationCancelRequest struct {
	Token string `form:"token" binding:"required,len=44"`
}
type ConfirmationCancelResponse struct {
	ClassType string `json:"class_type"`
	Date      string `json:"date"`
	Hour      string `json:"hour"`
	Location  string `json:"location"`
}

// TODO: do I want to return all this in cancel response?
func ToConfirmationCancelResponse(class models.Class) ConfirmationCancelResponse {
	return ConfirmationCancelResponse{
		ClassType: class.ClassCategory,
		Date:      fmt.Sprintf("%d %s %d", class.StartTime.Day(), class.StartTime.Month(), class.StartTime.Year()),
		Hour:      fmt.Sprintf("%d:%02d", class.StartTime.Hour(), class.StartTime.Minute()),
		Location:  class.Location,
	}
}

type ConfirmationCreateRequest struct {
	Token string `form:"token" binding:"required,len=44"`
}
type ConfirmationCreateResponse struct {
	ClassType string `json:"class_type"`
	Date      string `json:"date"`
	Hour      string `json:"hour"`
	Location  string `json:"location"`
}

func ToConfirmationCreateResponse(class models.Class) ConfirmationCreateResponse {
	return ConfirmationCreateResponse{
		ClassType: class.ClassCategory,
		Date:      fmt.Sprintf("%d %s %d", class.StartTime.Day(), class.StartTime.Month(), class.StartTime.Year()),
		Hour:      fmt.Sprintf("%d:%02d", class.StartTime.Hour(), class.StartTime.Minute()),
		Location:  class.Location,
	}
}

// TODO: w routach pending jest classID - ale nie jest parsowane z URL - zobacz co lepsze
type PendingOperationCreateRequest struct {
	ClassID   string `form:"class_id" binding:"required,uuid"`
	FirstName string `form:"first_name" binding:"required,min=3,max=30"`
	LastName  string `form:"last_name" binding:"required,max=30"`
	Email     string `form:"email" binding:"required,email"`
}

type PendingOperationCreateResponse struct {
	ClassID uuid.UUID `json:"class_id"`
}

type PendingOperationCancelRequest struct {
	ClassID   string `form:"class_id" binding:"required,uuid"`
	FirstName string `form:"first_name" binding:"required,min=3,max=30"`
	Email     string `form:"email" binding:"required,email"`
}

type PendingOperationCancelResponse struct {
	ClassID uuid.UUID `json:"class_id"`
}
