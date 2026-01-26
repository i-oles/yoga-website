package errs

import (
	"errors"
)

var ErrNoRowsAffected = errors.New("no rows affected")

var ErrAlreadyExist = errors.New("already exist in database")

var ErrNotFound = errors.New("not found in database")
