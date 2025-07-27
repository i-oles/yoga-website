package errs

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorHandler interface {
	Handle(c *gin.Context, tmplName string, err error)
}

type ErrorResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

type BookingError struct {
	Code    int
	Message string
	Err     error
}

func ErrAlreadyBooked(email string) *BookingError {
	return &BookingError{Code: http.StatusConflict, Message: "booking for email: " + email + " already exists in this class"}
}

func ErrClassFullyBooked(err error) *BookingError {
	return &BookingError{Code: http.StatusConflict, Message: "this class is fully booked, please choose different term", Err: err}
}

func (e *BookingError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}

	return e.Message
}
