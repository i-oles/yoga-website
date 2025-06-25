package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"main/configuration"
	"main/internal/repository/postgres"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		slog.Error("failed to loading configuration", slog.String("err", err.Error()))
		os.Exit(1)
	}

	connStr := fmt.Sprintf("dbname=%s user=%s password=%s host=localhost sslmode=disable",
		cfg.PostgresDBName,
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	slog.Info("Successfully connected to database")

	repo := postgres.NewClasses(db)
	resp, err := repo.GetCurrentMonthClasses()
	if err != nil {
		log.Fatal(err.Error())
	}

	fmt.Println(resp)
}

func loadConfig() (configuration.Configuration, error) {
	err := godotenv.Load()
	if err != nil {
		slog.Info("No .env file found, using environment variables...")
	}

	postgresUser := os.Getenv("POSTGRES_USER")
	postgresPassword := os.Getenv("POSTGRES_PASSWORD")
	if postgresUser == "" || postgresPassword == "" {
		return configuration.Configuration{},
			errors.New("provide envs for postgres access")
	}

	var cfg configuration.Configuration

	err = configuration.GetConfig("./config", &cfg)
	if err != nil {
		return configuration.Configuration{},
			fmt.Errorf("error loading configuration: %w", err)
	}

	slog.Info(cfg.Pretty())

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	return cfg, nil
}
