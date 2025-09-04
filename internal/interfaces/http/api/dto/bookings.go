package dto

import (
	"fmt"
	domainModels "main/internal/domain/models"
	"main/pkg/converter"
	"time"

	"github.com/google/uuid"
)

type GetAllBookingsForClassRequest struct {
	ClassID string `uri:"class_id" binding:"required"`
}

type BookingResponse struct {
	ID        uuid.UUID `json:"id"`
	ClassID   uuid.UUID `json:"class_id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func ToBookingResponse(booking domainModels.Booking) (BookingResponse, error) {
	createdAtWarsawTime, err := converter.ConvertToWarsawTime(booking.CreatedAt)
	if err != nil {
		return BookingResponse{}, fmt.Errorf("could not convert createdAt to warsaw time: %w", err)
	}

	return BookingResponse{
		ID:        booking.ID,
		ClassID:   booking.ClassID,
		FirstName: booking.FirstName,
		LastName:  booking.LastName,
		Email:     booking.Email,
		CreatedAt: createdAtWarsawTime,
	}, nil
}

func ToBookingsListResponse(bookings []domainModels.Booking) ([]BookingResponse, error) {
	bookingsListResponse := make([]BookingResponse, len(bookings))
	for i, booking := range bookings {
		bookingResponse, err := ToBookingResponse(booking)
		if err != nil {
			return nil, fmt.Errorf("could not convert booking to bookingResponse: %w", err)
		}

		bookingsListResponse[i] = bookingResponse
	}

	return bookingsListResponse, nil
}
