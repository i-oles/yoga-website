package dto

import (
	"fmt"
	"main/internal/domain/models"
	"main/pkg/converter"

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
	StartDate       string     `json:"start_date"`
	StartHour       string     `json:"start_hour"`
	ClassLevel      ClassLevel `json:"class_level"`
	ClassCategory   string     `json:"class_category"`
	CurrentCapacity int        `json:"current_capacity"`
	MaxCapacity     int        `json:"max_capacity"`
	Location        string     `json:"location"`
}

func ToClassResponse(class models.Class) (ClassResponse, error) {
	warsawTime, err := converter.ConvertToWarsawTime(class.StartTime)
	if err != nil {
		return ClassResponse{}, fmt.Errorf("error while converting time to warsaw time: %w", err)
	}

	return ClassResponse{
		ID:              class.ID,
		DayOfWeek:       class.DayOfWeek,
		StartDate:       warsawTime.Format(converter.DateLayout),
		StartHour:       warsawTime.Format(converter.HourLayout),
		ClassLevel:      ClassLevel(class.ClassLevel),
		ClassCategory:   class.ClassCategory,
		CurrentCapacity: class.CurrentCapacity,
		MaxCapacity:     class.MaxCapacity,
		Location:        class.Location,
	}, nil
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

func ToConfirmationCancelResponse(class models.Class) (ConfirmationCancelResponse, error) {
	warsawTimeDate, err := converter.ConvertToWarsawTime(class.StartTime)
	if err != nil {
		return ConfirmationCancelResponse{}, fmt.Errorf("could not convert start time: %w", err)
	}

	return ConfirmationCancelResponse{
		ClassType: class.ClassCategory,
		Date:      warsawTimeDate.Format(converter.DateLayout),
		Hour:      warsawTimeDate.Format(converter.HourLayout),
		Location:  class.Location,
	}, nil
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

func ToConfirmationCreateResponse(class models.Class) (ConfirmationCreateResponse, error) {
	warsawTimeDate, err := converter.ConvertToWarsawTime(class.StartTime)
	if err != nil {
		return ConfirmationCreateResponse{}, fmt.Errorf("could not convert start time: %w", err)
	}

	return ConfirmationCreateResponse{
		ClassType: class.ClassCategory,
		Date:      warsawTimeDate.Format(converter.DateLayout),
		Hour:      warsawTimeDate.Format(converter.HourLayout),
		Location:  class.Location,
	}, nil
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
	ClassID string `form:"class_id" binding:"required,uuid"`
	Email   string `form:"email" binding:"required,email"`
}

type PendingOperationCancelResponse struct {
	ClassID uuid.UUID `json:"class_id"`
}
