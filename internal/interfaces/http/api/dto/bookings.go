package dto

import (
	"fmt"
	"time"

	domainModels "main/internal/domain/models"
	"main/internal/interfaces/http/shared/dto"
	"main/pkg/converter"

	"github.com/google/uuid"
)

type BookingResponse struct {
	ID        uuid.UUID     `json:"id"`
	ClassID   uuid.UUID     `json:"class_id"`
	FirstName string        `json:"first_name"`
	LastName  string        `json:"last_name"`
	Email     string        `json:"email"`
	CreatedAt time.Time     `json:"created_at"`
	Class     *dto.ClassDTO `json:"class,omitempty"`
}

func ToBookingResponse(booking domainModels.Booking) (BookingResponse, error) {
	createdAtWarsawTime, err := converter.ConvertToWarsawTime(booking.CreatedAt)
	if err != nil {
		return BookingResponse{}, fmt.Errorf("could not convert createdAt to warsaw time: %w", err)
	}

	resp := BookingResponse{
		ID:        booking.ID,
		ClassID:   booking.ClassID,
		FirstName: booking.FirstName,
		LastName:  booking.LastName,
		Email:     booking.Email,
		CreatedAt: createdAtWarsawTime,
	}

	if booking.Class != nil {
		class, err := dto.ToClassDTO(*booking.Class)
		if err != nil {
			return BookingResponse{}, fmt.Errorf("could not cast class to dto class: %w", err)
		}

		resp.Class = &class
	}

	return resp, nil
}

func ToBookingsListResponse(bookings []domainModels.Booking) ([]BookingResponse, error) {
	bookingsListResponse := make([]BookingResponse, len(bookings))

	for idx, booking := range bookings {
		bookingResponse, err := ToBookingResponse(booking)
		if err != nil {
			return nil, fmt.Errorf("could not convert booking to bookingResponse: %w", err)
		}

		bookingsListResponse[idx] = bookingResponse
	}

	return bookingsListResponse, nil
}
