package view

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
	InvalidCancellationLinkCode
)

type ViewError struct {
	ClassID *uuid.UUID
	Code    int
	Message string
	Err     error
}

func (e *ViewError) Error() string {
	return e.Err.Error()
}

func ErrBookingAlreadyExists(classID uuid.UUID, email string, err error) *ViewError {
	return &ViewError{
		ClassID: &classID,
		Code:    BookingAlreadyExistsCode,
		Message: "Wygląda na to, że rezerwacja dla: " + email + " już istnieje. " +
			"Sprawdź skrzynkę mailową, aby znaleźć wcześniejsze potwierdzenie.",
		Err: err,
	}
}

func ErrBookingNotFound(classID uuid.UUID, email string, err error) *ViewError {
	return &ViewError{
		ClassID: &classID,
		Code:    BookingNotFoundCode,
		Message: "Brak potwierdzonej rezerwacji na te zajęcia dla: " + email,
		Err:     err,
	}
}

func ErrClassExpired(classID uuid.UUID, err error) *ViewError {
	return &ViewError{
		ClassID: &classID,
		Code:    ClassExpiredCode,
		Message: "Rezerwacja niedostępna – zajęcia już się zaczęły albo odbyły.",
		Err:     err,
	}
}

func ErrPendingBookingNotFound(err error) *ViewError {
	return &ViewError{
		Code:    PendingBookingNotFoundCode,
		Message: "Link potwierdzający rezerwację wygasł, rozpocznij nową rezerwację.",
		Err:     err,
	}
}

func ErrTooManyPendingBookings(classID uuid.UUID, email string, err error) *ViewError {
	return &ViewError{
		Code:    TooManyPendingBookingsCode,
		ClassID: &classID,
		Message: fmt.Sprintf("Wyczerpano limit linków potwierdzających dla %s. "+
			"Sprawdź wiadości odebrane lub spam w skrzynce mailowej.", email),
		Err: err,
	}
}

func ErrClassFullyBooked(classID uuid.UUID, err error) *ViewError {
	return &ViewError{
		Code:    ClassFullyBookedCode,
		ClassID: &classID,
		Message: "Brak wolnych miejsc na te zajęcia",
		Err:     err,
	}
}

func ErrSomeoneBookedClassFaster(err error) *ViewError {
	return &ViewError{
		Code:    SomeoneBookedClassFasterCode,
		Message: "Ktoś Cię uprzedził... :( Brak wolnych miejsc na te zajęcia.",
		Err:     err,
	}
}

func ErrInvalidCancellationLink(err error) *ViewError {
	return &ViewError{
		Code:    InvalidCancellationLinkCode,
		Message: "Link do odwołania rezerwacji wygasł albo jest nieprawidłowy, skontaktuj się ze mną.",
		Err:     err,
	}
}
