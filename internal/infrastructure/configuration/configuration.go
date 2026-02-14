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

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("could not unmarshal duration string: %w", err)
	}

	parsed, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("could not parse duration: %w", err)
	}

	d.Duration = parsed

	return nil
}

type EmailSenderSettings struct {
	Host     string
	Port     int
	User     string
	Password string
	FromName string
}

type Configuration struct {
	ListenAddress                    string
	DBPath                           string
	ReadTimeout                      Duration
	WriteTimeout                     Duration
	ContextTimeout                   Duration
	AuthSecret                       string
	LogBusinessErrors                bool
	LogConfig                        bool
	EmailSender                      EmailSenderSettings
	DomainAddr                       string
	ConfirmationRequestEmailTmplPath string
	ConfirmationEmailTmplPath        string
	BaseSenderTmplPath               string
	IsVacation                       bool
}

func (c *Configuration) Pretty() string {
	cfgPretty, _ := json.MarshalIndent(c, "", "  ")

	return string(cfgPretty)
}

func GetConfig(cfgPath string) (*Configuration, error) {
	cfg := &Configuration{}

	err := godotenv.Load()
	if err != nil {
		slog.Info("No .env file found, using environment variables...")
	}

	configFileName := os.Getenv("CONFIG")

	cfgFinalPath := filepath.Join(cfgPath, configFileName+".json")

	err = gonfig.GetConf(cfgFinalPath, cfg)
	if err != nil {
		return nil, fmt.Errorf("could not load configuration: %s", err.Error())
	}

	loadEnvs(cfg)

	if cfg.EmailSender.User == "" || cfg.EmailSender.Password == "" {
		return nil,
			errors.New("provide envs for email sender")
	}

	return cfg, nil
}

func loadEnvs(cfg *Configuration) {
	if emailSenderUser := os.Getenv("EMAIL_SENDER_USER"); emailSenderUser != "" {
		cfg.EmailSender.User = emailSenderUser
	}

	if emailSenderPassword := os.Getenv("EMAIL_SENDER_PASSWORD"); emailSenderPassword != "" {
		cfg.EmailSender.Password = emailSenderPassword
	}

	if authSecret := os.Getenv("AUTH_SECRET"); authSecret != "" {
		cfg.AuthSecret = authSecret
	}

	if dbPath := os.Getenv("DATABASE_PATH"); dbPath != "" {
		cfg.DBPath = dbPath
	}
}
