package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"main/internal/domain/models"
	"main/internal/infrastructure/configuration"
	"main/internal/infrastructure/notifier/gmail"

	"github.com/google/uuid"
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		slog.Error("failed to loading configuration", slog.String("err", err.Error()))
		os.Exit(1)
	}

	notifier := gmail.NewNotifier(
		cfg.Notifier.Host,
		cfg.Notifier.Port,
		cfg.Notifier.Signature,
		cfg.Notifier.Login,
		cfg.Notifier.Password,
		cfg.BaseNotifierTmplPath,
	)

	four := 4

	notifierParams := models.NotifierParams{
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

	err = notifier.NotifyBookingConfirmation(notifierParams, cancellationLink)
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
