package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"main/internal/domain/repositories"
	"main/internal/infrastructure/configuration"
	dbModels "main/internal/infrastructure/models/db"
	sqliteRepo "main/internal/infrastructure/repository/sqlite"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Components struct {
	bookingsRepo repositories.IBookings
	passesRepo   repositories.IPasses
	database     *gorm.DB
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		slog.Error("failed to loading configuration", slog.String("err", err.Error()))
		os.Exit(1)
	}

	components, err := buildComponents(cfg)
	if err != nil {
		slog.Error("failed to build components", slog.String("err", err.Error()))
		os.Exit(1)
	}

	bookingsRepo := components.bookingsRepo
	passRepo := components.passesRepo

	ctx := context.Background()

	type user struct {
		email      string
		bookingIDs []uuid.UUID
	}

	users := []user{
		{
			email:      "oles.robert@gmail.com",
			bookingIDs: []uuid.UUID{},
		},
		{
			email: "paulina.sergiel@gmail.com",
			bookingIDs: []uuid.UUID{
				uuid.MustParse("b1224afa-c5f4-439c-8e1c-188c53bd4e5a"),
				uuid.MustParse("2cf89329-1152-44b3-956d-e6e404279edc"),
				uuid.MustParse("5c7c3a3c-f749-4126-bed4-83aa6dc2bc81"),
			},
		},
		{
			email: "listyzebrane@gmail.com",
			bookingIDs: []uuid.UUID{
				uuid.MustParse("19ce0e63-f014-41ab-8307-5f8a9ad8a817"),
				uuid.MustParse("b94798a1-8289-448d-ad49-bc301456c2a1"),
				uuid.MustParse("97fbc3d5-ccf7-4299-95fe-7a90fb153e22"),
				uuid.MustParse("09cb76fc-24dc-43d1-8023-ca87fd5672a7"),
				uuid.MustParse("2e0eef71-fb56-400e-ae4b-85f8d4430e95"),
			},
		},
		{
			email: "mira.rismiatova@gmail.com",
			bookingIDs: []uuid.UUID{
				uuid.MustParse("99a1a188-d311-4b84-92ab-feaf5f252d4d"),
			},
		},
	}

	var count int64

	if err := components.database.
		Model(&dbModels.SQLPass{}).
		Count(&count).Error; err != nil {
		slog.Error(
			"failed to count passes",
			slog.String("err", err.Error()),
		)
		os.Exit(1)
	}

	if count > 0 {
		slog.Error(
			"passes table is not empty, aborting migration",
			slog.Int64("passes_count", count),
		)
		os.Exit(1)
	}

	slog.Info("starting passes migration")

	err = components.database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		bookingsRepo := sqliteRepo.NewBookingsRepo(tx)
		passesRepo := sqliteRepo.NewPassesRepo(tx)

		for _, user := range users {
			slog.Info(
				"creating pass",
				slog.String("email", user.email),
				slog.Int("used_slots", len(user.bookingIDs)),
			)

			pass, err := passesRepo.Insert(ctx, user.email, 5)
			if err != nil {
				return fmt.Errorf(
					"failed to create pass for %s: %w",
					user.email,
					err,
				)
			}

			slog.Info(
				"pass created",
				slog.String("email", user.email),
				slog.Int("pass_id", pass.ID),
			)

			for _, bookingID := range user.bookingIDs {
				slog.Info(
					"assigning booking to pass",
					slog.String("email", user.email),
					slog.Int("pass_id", pass.ID),
					slog.String("booking_id", bookingID.String()),
				)

				err := bookingsRepo.Update(ctx, bookingID, map[string]any{
					"pass_id": pass.ID,
				})
				if err != nil {
					return fmt.Errorf(
						"failed to assign booking %s to pass %d: %w",
						bookingID,
						pass.ID,
						err,
					)
				}
			}
		}

		return nil
	})
	if err != nil {
		slog.Error(
			"migration failed - transaction rolled back",
			slog.String("err", err.Error()),
		)
		os.Exit(1)
	}

	slog.Info(
		"migration completed successfully",
		slog.Int("users", len(users)),
	)
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

func buildComponents(cfg *configuration.Configuration) (Components, error) {
	database, err := gorm.Open(sqlite.Open(cfg.DBPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return Components{}, fmt.Errorf("failed to connect to database: %w", err)
	}

	slog.Info("Successfully connected to database")

	err = database.AutoMigrate(
		&dbModels.SQLClass{},
		&dbModels.SQLPendingBooking{},
		&dbModels.SQLBooking{},
		&dbModels.SQLPass{},
	)
	if err != nil {
		return Components{}, fmt.Errorf("failed to migrate database: %w", err)
	}

	bookingsRepo := sqliteRepo.NewBookingsRepo(database)
	passesRepo := sqliteRepo.NewPassesRepo(database)

	return Components{
		bookingsRepo: bookingsRepo,
		passesRepo:   passesRepo,
		database:     database,
	}, nil
}
