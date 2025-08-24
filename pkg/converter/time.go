package converter

import (
	"fmt"
	"time"
)

const (
	// DateLayout should not be changed, it can cause an error
	DateLayout = "02-01-2006"
	HourLayout = "15:04"
)

func ConvertToWarsawTime(t time.Time) (time.Time, error) {
	loc, err := time.LoadLocation("Europe/Warsaw")
	if err != nil {
		return time.Time{}, fmt.Errorf("error while loading location: %w", err)
	}

	return t.In(loc), nil
}
