package errs

import (
	"fmt"

	"github.com/google/uuid"
)

// TODO: should error messages be in english?

const (
	ConfirmedBookingNotFoundCode int = iota
	ConfirmedBookingAlreadyExistsCode
	ExpiredClassBookingCode
	PendingOperationNotFoundCode
	TooManyPendingOperationsCode
	ClassFullyBookedCode
	ClassEmptyCode
	SomeoneBookedClassFasterCode
)

func ErrConfirmedBookingAlreadyExists(email string, err error) *BookingError {
	return &BookingError{
		Code: ConfirmedBookingAlreadyExistsCode,
		Message: "Wygląda na to, że rezerwacja dla: " + email + " już istnieje. " +
			"Sprawdź skrzynkę mailową, aby znaleźć wcześniejsze potwierdzenie.",
		Err: err,
	}
}

func ErrConfirmedBookingNotFound(classID uuid.UUID, email string, err error) *BookingError {
	return &BookingError{
		ClassID: &classID,
		Code:    ConfirmedBookingNotFoundCode,
		Message: "Brak potwierdzonej rezerwacji na te zajęcia dla: " + email,
		Err:     err,
	}
}

func ErrExpiredClassBooking(classID uuid.UUID, err error) *BookingError {
	return &BookingError{
		ClassID: &classID,
		Code:    ExpiredClassBookingCode,
		Message: "Rezerwacja niedostępna – zajęcia już się zaczęły albo odbyły",
		Err:     err,
	}
}

func ErrPendingOperationNotFound(err error) *BookingError {
	return &BookingError{
		Code:    PendingOperationNotFoundCode,
		Message: "Twój link potwierdzający operację wygasł, stwórz nową sesję.",
		Err:     err,
	}
}

func ErrTooManyPendingOperations(classID uuid.UUID, email string, err error) *BookingError {
	return &BookingError{
		Code:    TooManyPendingOperationsCode,
		ClassID: &classID,
		Message: fmt.Sprintf("Wyczerpano limit linków potwierdzających dla %s. "+
			"Sprawdź wiadości odebrane lub spam w skrzynce mailowej.", email),
		Err: err,
	}
}

func ErrClassFullyBooked(classID uuid.UUID, err error) *BookingError {
	return &BookingError{
		Code:    ClassFullyBookedCode,
		ClassID: &classID,
		Message: "Brak wolnych miejsc na te zajęcia",
		Err:     err,
	}
}

func ErrSomeoneBookedClassFaster(err error) *BookingError {
	return &BookingError{
		Code:    SomeoneBookedClassFasterCode,
		Message: "Ktoś Cię uprzedził... :( Brak wolnych miejsc na te zajęcia.",
		Err:     err,
	}
}

func ErrClassEmpty(classID uuid.UUID, err error) *BookingError {
	return &BookingError{
		Code:    ClassEmptyCode,
		ClassID: &classID,
		Message: "Nie możesz odwołać zajęć, ponieważ nikt jeszcze nie zrobił rezerwacji.",
		Err:     err,
	}
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
