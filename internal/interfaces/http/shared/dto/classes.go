package dto

import (
	"fmt"
	"main/internal/domain/models"
	"main/pkg/converter"
	"main/pkg/translator"

	"github.com/google/uuid"
)

type ClassDTO struct {
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

func ToClassDTO(class models.Class) (ClassDTO, error) {
	warsawTime, err := converter.ConvertToWarsawTime(class.StartTime)
	if err != nil {
		return ClassDTO{}, fmt.Errorf("error while converting time to warsaw time: %w", err)
	}

	weekday, err := translator.TranslateToWeekDayToPolish(warsawTime.Weekday())
	if err != nil {
		return ClassDTO{}, fmt.Errorf("error while translating week day to polish: %w", err)
	}

	return ClassDTO{
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

func ToClassesListDTO(classes []models.Class) ([]ClassDTO, error) {
	classesResponse := make([]ClassDTO, len(classes))
	for i, class := range classes {
		classResponse, err := ToClassDTO(class)
		if err != nil {
			return nil, fmt.Errorf("could not convert class to classResponse: %w", err)
		}

		classesResponse[i] = classResponse
	}

	return classesResponse, nil
}
