package services

import (
	"sort"
	"time"

	"main/internal/domain/models"
)

type PassManager struct{}

func (p *PassManager) BuildPassItems(
	bookings []models.Booking,
	totalBookings int,
) []models.PassItem {
	passItems := make([]models.PassItem, 0, totalBookings)

	for _, booking := range bookings {
		classStartTime := booking.Class.StartTime
		passItem := models.PassItem{
			ClassStartTime: &classStartTime,
		}

		if classStartTime.Before(time.Now()) {
			passItem.Status = models.PastPassStatus
		} else {
			passItem.Status = models.FuturePassStatus
		}

		passItems = append(passItems, passItem)
	}

	if len(passItems) < totalBookings {
		for i := len(passItems); i < totalBookings; i++ {
			passItems = append(passItems, models.PassItem{
				Status: models.BlankPassStatus,
			})
		}
	}

	sort.Slice(passItems, func(i, j int) bool {
		a := passItems[i]
		b := passItems[j]

		if a.Status == models.BlankPassStatus && b.Status != models.BlankPassStatus {
			return false
		}

		if a.Status != models.BlankPassStatus && b.Status == models.BlankPassStatus {
			return true
		}

		if a.ClassStartTime != nil && b.ClassStartTime != nil {
			return a.ClassStartTime.Before(*b.ClassStartTime)
		}

		return false
	})

	return passItems
}
