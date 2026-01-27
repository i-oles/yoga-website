package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"main/internal/domain/models"
	"main/internal/infrastructure/configuration"
	"main/internal/infrastructure/sender/gmail"
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		slog.Error("failed to loading configuration", slog.String("err", err.Error()))
		os.Exit(1)
	}

	emailSender := gmail.NewSender(
		cfg.EmailSender.Host,
		cfg.EmailSender.Port,
		cfg.EmailSender.User,
		cfg.EmailSender.Password,
		cfg.EmailSender.FromName,
		cfg.BaseSenderTmplPath,
	)

	three := 3
	four := 4

	senderParams := models.SenderParams{
		RecipientEmail:     "orth.quala@gmail.com",
		RecipientFirstName: "orth",
		ClassName:          "vinyasa",
		ClassLevel:         "beginner",
		StartTime:          time.Now(),
		Location:           "dom",
		UsedPassCredits:    *three,
		TotalPassCredits:   *four,
	}

	cancellationLink := "http://testlink.com"

	err = emailSender.SendConfirmations(senderParams, cancellationLink)
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
