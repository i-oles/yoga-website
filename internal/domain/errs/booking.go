package errs

import (
	"fmt"

	"github.com/google/uuid"
)

const (
	BookingNotFoundCode int = iota
	BookingAlreadyExistsCode
	ClassExpiredCode
	PendingBookingNotFoundCode
	TooManyPendingBookingsCode
	ClassFullyBookedCode
	ClassEmptyCode
	SomeoneBookedClassFasterCode
)

type BookingError struct {
	ClassID *uuid.UUID
	Code    int
	Message string
	Err     error
}

func (e *BookingError) Error() string {
	return e.Err.Error()
}

func ErrBookingAlreadyExists(classID uuid.UUID, email string, err error) *BookingError {
	return &BookingError{
		ClassID: &classID,
		Code:    BookingAlreadyExistsCode,
		Message: "Wygląda na to, że rezerwacja dla: " + email + " już istnieje. " +
			"Sprawdź skrzynkę mailową, aby znaleźć wcześniejsze potwierdzenie.",
		Err: err,
	}
}

func ErrBookingNotFound(classID uuid.UUID, email string, err error) *BookingError {
	return &BookingError{
		ClassID: &classID,
		Code:    BookingNotFoundCode,
		Message: "Brak potwierdzonej rezerwacji na te zajęcia dla: " + email,
		Err:     err,
	}
}

func ErrClassExpired(classID uuid.UUID, err error) *BookingError {
	return &BookingError{
		ClassID: &classID,
		Code:    ClassExpiredCode,
		Message: "Rezerwacja niedostępna – zajęcia już się zaczęły albo odbyły",
		Err:     err,
	}
}

func ErrPendingBookingNotFound(err error) *BookingError {
	return &BookingError{
		Code:    PendingBookingNotFoundCode,
		Message: "Twój link potwierdzający operację wygasł, stwórz nową sesję.",
		Err:     err,
	}
}

func ErrTooManyPendingBookings(classID uuid.UUID, email string, err error) *BookingError {
	return &BookingError{
		Code:    TooManyPendingBookingsCode,
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
