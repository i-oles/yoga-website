package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"main/internal/domain/models"
	"main/internal/infrastructure/configuration"
	"main/internal/interfaces/http/html/dto"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		slog.Error("failed to loading configuration", slog.String("err", err.Error()))
		os.Exit(1)
	}

	router := gin.Default()

	router.Static("web/static", "./web/static")
	router.LoadHTMLGlob("web/templates/*")
	api := router.Group("/")

	api.GET("/", preview)

	srv := &http.Server{
		Addr:              ":8000",
		Handler:           router,
		ReadHeaderTimeout: cfg.ReadTimeout * time.Second,
		ReadTimeout:       cfg.ReadTimeout * time.Second,
		WriteTimeout:      cfg.WriteTimeout * time.Second,
	}

	runServer(srv, cfg)
}

func preview(c *gin.Context) {
	class := models.Class{
		ID:          uuid.New(),
		StartTime:   time.Now(),
		ClassLevel:  "Beginner",
		ClassName:   "Vinyasa",
		MaxCapacity: 5,
		Location:    "Studio A",
	}

	view, err := dto.ToClassView(class)
	if err != nil {
		fmt.Println(err.Error())

		return
	}

	// c.HTML(http.StatusOK, "confirmation_create_booking.tmpl", view)

	c.HTML(http.StatusOK, "cancel_booking_form.tmpl", gin.H{
		"Class": view, "BookingID": uuid.New(), "ConfirmationToken": "abrakadabra",
	})
}

func runServer(srv *http.Server, cfg *configuration.Configuration) {
	go func() {
		slog.Info("Starting server...", slog.String("address", ":8000"))

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
