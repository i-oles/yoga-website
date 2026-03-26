package main

import (
	"context"
	"errors"
	"log/slog"
	"main/internal/domain/notifier"
	"main/internal/domain/repositories"
	"time"

	"gorm.io/gorm"
)

type classReminder struct {
	classesRepo  repositories.IClasses
	bookingsRepo repositories.IBookings
	notifier     notifier.INotifier
}

func (c classReminder) NewClassReminder(
	classesRepo repositories.IClasses,
	bookingsRepo repositories.IBookings,
	notifier notifier.INotifier,
) classReminder {
	return classReminder{
		classesRepo:  classesRepo,
		bookingsRepo: bookingsRepo,
		notifier:     notifier,
	}
}

func (c classReminder) Remind()  {
		//nolint
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		tenHoursAgo := time.Now().UTC().Add(-10 * time.Hour)

		var pendingBooking dbModels.SQLPendingBooking

		result := database.WithContext(ctx).
			Where("created_at < ?", oneHourAgo).
			Delete(&pendingBooking)

		if result.Error != nil {
			if errors.Is(result.Error, context.DeadlineExceeded) {
				slog.Warn("cleanup timeout exceeded")
			} else {
				slog.Error("failed to cleanup pending bookings async",
					slog.String("err", result.Error.Error()))
			}

			return
		}

		slog.Info("Cleaned up pending bookings", slog.Int64("rows_deleted", result.RowsAffected))
	}()
}
