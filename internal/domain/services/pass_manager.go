package services

import (
	"sort"
	"time"

	"main/internal/domain/models"
)

type PassManager struct{}

func (p *PassManager) BuildPassSlots(
	bookings []models.Booking,
	totalSlots int,
) []models.PassSlot {
	passSlots := make([]models.PassSlot, 0, totalSlots)

	for _, booking := range bookings {
		classStartTime := booking.Class.StartTime
		passSlot := models.PassSlot{
			ClassStartTime: &classStartTime,
		}

		if classStartTime.Before(time.Now()) {
			passSlot.Status = models.PastPassStatus
		} else {
			passSlot.Status = models.FuturePassStatus
		}

		passSlots = append(passSlots, passSlot)
	}

	if len(passSlots) < totalSlots {
		for i := len(passSlots); i < totalSlots; i++ {
			passSlots = append(passSlots, models.PassSlot{
				Status: models.BlankPassStatus,
			})
		}
	}

	sort.Slice(passSlots, func(i, j int) bool {
		a := passSlots[i]
		b := passSlots[j]

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

	return passSlots
}
