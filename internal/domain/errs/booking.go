package errs

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

const (
	ConfirmedBookingNotFoundCode int = iota
	ConfirmedBookingAlreadyExistsCode
)

func ErrConfirmedBookingAlreadyExists(email string) *BookingError {
	return &BookingError{
		Code: ConfirmedBookingAlreadyExistsCode,
		// TODO: should this be in english?
		Message: "Wygląda na to, że rezerwacja dla: " + email + " już istnieje. " +
			"Sprawdź swoją skrzynkę pocztową, aby znaleźć wcześniejsze potwierdzenie.",
	}
}

func ErrConfirmedBookingNotFound(email string, classID uuid.UUID) *BookingError {
	return &BookingError{
		ClassID: &classID,
		Code:    ConfirmedBookingNotFoundCode,
		Message: "Brak potwierdzonej rezerwacji na te zajęcia dla: " + email,
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
