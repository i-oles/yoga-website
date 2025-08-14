package errs

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

const (
	ConfirmedBookingNotFoundCode int = iota
)

func ErrConfirmedBookingAlreadyExists(email string) *BookingError {
	return &BookingError{Code: http.StatusConflict, Message: "booking for : " + email + " already exists in this class"}
}

func ErrConfirmedBookingNotFound(email string, classID uuid.UUID) *BookingError {
	return &BookingError{
		ClassID: &classID,
		Code:    ConfirmedBookingNotFoundCode,
		Message: "confirmed booking for " + email + " does not exist in this class",
	}
}

func ErrPendingOperationCreateAlreadyExists(email string) *BookingError {
	return &BookingError{Code: http.StatusConflict, Message: "link for booking confirmation was already sent to : " + email}
}

func ErrClassFullyBooked(err error) *BookingError {
	return &BookingError{Code: http.StatusConflict, Message: "this class is fully booked, please choose different term", Err: err}
}

func ErrClassEmpty(err error) *BookingError {
	return &BookingError{Code: http.StatusConflict, Message: "this class is empty, booking cancellation is impossible", Err: err}
}

type BookingError struct {
	ClassID *uuid.UUID
	Code    int
	Message string
	Err     error
}

func (e *BookingError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}

	return e.Message
}
