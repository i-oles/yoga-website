package configuration

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	"github.com/tkanos/gonfig"
)

const (
	defaultCfgFilename = "dev.json"
)

type EmailSenderSettings struct {
	Host     string
	Port     int
	User     string
	Password string
	FromName string
}

type PostgresSettings struct {
	User     string
	Password string
	DBName   string
}

type Configuration struct {
	ListenAddress  string
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	ContextTimeout time.Duration
	Postgres       PostgresSettings
	LogErrors      bool
	EmailSender    EmailSenderSettings
}

func (c *Configuration) Pretty() string {
	cfgPretty, _ := json.MarshalIndent(c, "", "  ")

	return string(cfgPretty)
}

func GetConfig(cfgPath string) (*Configuration, error) {
	cfg := &Configuration{}

	cfgFinalPath := filepath.Join(cfgPath, defaultCfgFilename)

	err := gonfig.GetConf(cfgFinalPath, cfg)
	if err != nil {
		return nil, fmt.Errorf("could not load configuration: %s", err.Error())
	}

	loadEnvs(cfg)

	if cfg.Postgres.User == "" || cfg.Postgres.Password == "" {
		return nil,
			errors.New("provide envs for postgres access")
	}

	if cfg.EmailSender.User == "" || cfg.EmailSender.Password == "" {
		return nil,
			errors.New("provide envs for email sender")
	}

	return cfg, nil
}

func loadEnvs(cfg *Configuration) {
	err := godotenv.Load()
	if err != nil {
		slog.Info("No .env file found, using environment variables...")
	}

	if user := os.Getenv("EMAIL_SENDER_USER"); user != "" {
		cfg.EmailSender.User = user
	}

	if password := os.Getenv("EMAIL_SENDER_PASSWORD"); password != "" {
		cfg.EmailSender.Password = password
	}

	if postgresUser := os.Getenv("POSTGRES_USER"); postgresUser != "" {
		cfg.Postgres.User = postgresUser
	}

	if postgresPassword := os.Getenv("POSTGRES_PASSWORD"); postgresPassword != "" {
		cfg.Postgres.Password = postgresPassword
	}
}
