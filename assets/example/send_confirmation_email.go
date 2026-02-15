package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"main/internal/domain/models"
	"main/internal/infrastructure/configuration"
	"main/internal/infrastructure/sender/gmail"

	"github.com/google/uuid"
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		slog.Error("failed to loading configuration", slog.String("err", err.Error()))
		os.Exit(1)
	}

	emailSender := gmail.NewNotifier(
		cfg.Notifier.Host,
		cfg.Notifier.Port,
		cfg.Notifier.Signature,
		cfg.Notifier.Login,
		cfg.Notifier.Password,
		cfg.BaseSenderTmplPath,
	)

	four := 4

	senderParams := models.NotifierParams{
		RecipientEmail:     "orth.quala@gmail.com",
		RecipientFirstName: "orth",
		RecipientLastName:  "quala",
		ClassName:          "vinyasa",
		ClassLevel:         "beginner",
		StartTime:          time.Now(),
		Location:           "dom",
		PassUsedBookingIDs: []uuid.UUID{},
		PassTotalBookings:  &four,
	}

	cancellationLink := "http://testlink.com"

	err = emailSender.NotifyBookingConfirmation(senderParams, cancellationLink)
	if err != nil {
		slog.Error(err.Error())
	}
}

func loadConfig() (*configuration.Configuration, error) {
	cfg, err := configuration.GetConfig("./config")
	if err != nil {
		return nil, fmt.Errorf("error loading configuration: %w", err)
	}

	if cfg.LogConfig {
		slog.Info(cfg.Pretty())
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	return cfg, nil
}
