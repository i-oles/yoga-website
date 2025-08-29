package errs

import (
	"errors"
)

var ErrNoRowsAffected = errors.New("no rows affected")

var ErrNotFound = errors.New("not found in database")
