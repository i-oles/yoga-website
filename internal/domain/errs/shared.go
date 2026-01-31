package errs

import "errors"

var ErrBookingIDNotFoundInPass = errors.New("not found bookingID in pass.usedBookingIDs")
