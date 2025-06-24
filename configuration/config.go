package configuration

import (
	"encoding/json"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/tkanos/gonfig"
)

const (
	defaultCfgFilename = "dev.json"
)

type Configuration struct {
	ListenAddress  string
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	ContextTimeout time.Duration
	LogErrors      bool
}

func (c *Configuration) Pretty() string {
	cfgPretty, _ := json.MarshalIndent(c, "", "  ")

	return string(cfgPretty)
}

func GetConfig(cfgPath string, cfg *Configuration) error {
	cfgFinalPath := filepath.Join(cfgPath, defaultCfgFilename)

	err := gonfig.GetConf(cfgFinalPath, cfg)
	if err != nil {
		slog.Error("config error", slog.String("err", err.Error()))
	}

	return nil
}
