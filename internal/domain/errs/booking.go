package errs

import (
	"fmt"
	"net/http"
)

func ErrAlreadyBooked(email string) *BookingError {
	return &BookingError{Code: http.StatusConflict, Message: "confirmation for email: " + email + " already exists in this class"}
}

func ErrClassFullyBooked(err error) *BookingError {
	return &BookingError{Code: http.StatusConflict, Message: "this class is fully booked, please choose different term", Err: err}
}

func ErrClassEmpty(err error) *BookingError {
	return &BookingError{Code: http.StatusConflict, Message: "this class is empty, booking cancellation is impossible", Err: err}
}

type BookingError struct {
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
