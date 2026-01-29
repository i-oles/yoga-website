package tools

import (
	"main/internal/domain/errs"
)

func RemoveFromSlice[T comparable](slice []T, value T) ([]T, error) {
	for i, v := range slice {
		if v == value {
			return append(slice[:i], slice[i+1:]...), nil
		}
	}

	return nil, errs.ErrBookingIDNotFoundInPass
}
