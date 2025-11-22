package translator

import (
	"errors"
	"time"
)

func TranslateToWeekDayToPolish(weekDay time.Weekday) (string, error) {
	switch weekDay {
	case time.Monday:
		return "poniedziałek", nil
	case time.Tuesday:
		return "wtorek", nil
	case time.Wednesday:
		return "środa", nil
	case time.Thursday:
		return "czwartek", nil
	case time.Friday:
		return "piątek", nil
	case time.Saturday:
		return "sobota", nil
	case time.Sunday:
		return "niedziela", nil
	default:
		return "", errors.New("unknown weekday")
	}
}
