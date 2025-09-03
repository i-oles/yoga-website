package dto

import (
	"fmt"
	domainModels "main/internal/domain/models"
	"main/pkg/converter"
	"time"

	"github.com/google/uuid"
)

type ClassRequest struct {
	StartTime       time.Time `json:"start_time" binding:"required" time_format:"2006-01-02T15:04:05Z07:00"`
	ClassLevel      string    `json:"class_level" binding:"required,min=3,max=40"`
	ClassName       string    `json:"class_name" binding:"required,min=3,max=40"`
	CurrentCapacity int       `json:"current_capacity" binding:"gte=0"`
	MaxCapacity     int       `json:"max_capacity" binding:"gte=1"`
	Location        string    `json:"location" binding:"required"`
}

type ClassResponse struct {
	ID              uuid.UUID `json:"id"`
	WeekDay         string    `json:"week_day"`
	StartDate       string    `json:"start_date"`
	StartHour       string    `json:"start_hour"`
	ClassLevel      string    `json:"class_level"`
	ClassName       string    `json:"class_name"`
	CurrentCapacity int       `json:"current_capacity"`
	MaxCapacity     int       `json:"max_capacity"`
	Location        string    `json:"location"`
}

func ToClassResponse(class domainModels.Class) (ClassResponse, error) {
	warsawTime, err := converter.ConvertToWarsawTime(class.StartTime)
	if err != nil {
		return ClassResponse{}, fmt.Errorf("error while converting time to warsaw time: %w", err)
	}

	weekday, err := translateToWeekDayToPolish(warsawTime.Weekday())
	if err != nil {
		return ClassResponse{}, fmt.Errorf("error while translating week day to polish: %w", err)
	}

	return ClassResponse{
		ID:              class.ID,
		WeekDay:         weekday,
		StartDate:       warsawTime.Format(converter.DateLayout),
		StartHour:       warsawTime.Format(converter.HourLayout),
		ClassLevel:      class.ClassLevel,
		ClassName:       class.ClassName,
		CurrentCapacity: class.CurrentCapacity,
		MaxCapacity:     class.MaxCapacity,
		Location:        class.Location,
	}, nil
}

func ToClassesListResponse(classes []domainModels.Class) ([]ClassResponse, error) {
	classesResponse := make([]ClassResponse, len(classes))
	for i, class := range classes {
		classResponse, err := ToClassResponse(class)
		if err != nil {
			return nil, fmt.Errorf("could not convert class to classResponse: %w", err)
		}

		classesResponse[i] = classResponse
	}

	return classesResponse, nil
}

// TODO: create separate translator for this
func translateToWeekDayToPolish(weekDay time.Weekday) (string, error) {
	switch weekDay {
	case time.Monday:
		return "poniedziałek", nil
	case time.Tuesday:
		return "wtorek", nil
	case time.Wednesday:
		return "środa", nil
	case time.Thursday:
		return "czwartek", nil
	case time.Friday:
		return "piątek", nil
	case time.Saturday:
		return "sobota", nil
	case time.Sunday:
		return "niedziela", nil
	default:
		return "", fmt.Errorf("unknown weekday")
	}
}
