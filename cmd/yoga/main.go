package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"main/configuration"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		slog.Error("failed to loading configuration", slog.String("err", err.Error()))
		os.Exit(1)
	}

	router := setupRouter()

	srv := &http.Server{
		Addr:              cfg.ListenAddress,
		Handler:           router,
		ReadHeaderTimeout: cfg.ReadTimeout * time.Second,
		ReadTimeout:       cfg.ReadTimeout * time.Second,
		WriteTimeout:      cfg.WriteTimeout * time.Second,
	}

	runServer(srv, cfg)
}

func loadConfig() (configuration.Configuration, error) {
	var cfg configuration.Configuration

	err := configuration.GetConfig("./config", &cfg)
	if err != nil {
		return configuration.Configuration{},
			fmt.Errorf("error loading configuration: %w", err)
	}

	slog.Info(cfg.Pretty())

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	return cfg, nil
}

func setupRouter() *gin.Engine {
	router := gin.Default()

	_ = router.Group("/")

	return router
}

func runServer(srv *http.Server, cfg configuration.Configuration) {
	go func() {
		slog.Info("Starting server...", slog.String("address", cfg.ListenAddress))

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error: %s\n", slog.String("err", err.Error()))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	slog.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ContextTimeout*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", slog.String("err", err.Error()))
	}

	slog.Info("Server stopped")
}
