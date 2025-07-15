package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"main/configuration"
	"main/internal/generator/token"
	"main/internal/handler/book"
	"main/internal/handler/classes"
	"main/internal/handler/confirmation"
	"main/internal/handler/submit"
	"main/internal/repository/postgres"
	"main/internal/sender/email"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		slog.Error("failed to loading configuration", slog.String("err", err.Error()))
		os.Exit(1)
	}

	connStr := fmt.Sprintf("dbname=%s user=%s password=%s host=localhost sslmode=disable",
		cfg.Postgres.DBName,
		cfg.Postgres.User,
		cfg.Postgres.Password,
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

	router := setupRouter(db, cfg)

	srv := &http.Server{
		Addr:              cfg.ListenAddress,
		Handler:           router,
		ReadHeaderTimeout: cfg.ReadTimeout * time.Second,
		ReadTimeout:       cfg.ReadTimeout * time.Second,
		WriteTimeout:      cfg.WriteTimeout * time.Second,
	}

	runServer(srv, cfg)
}

func loadConfig() (*configuration.Configuration, error) {
	cfg, err := configuration.GetConfig("./config")
	if err != nil {
		return nil, fmt.Errorf("error loading configuration: %w", err)
	}

	slog.Info(cfg.Pretty())

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	return cfg, nil
}

func setupRouter(db *sql.DB, cfg *configuration.Configuration) *gin.Engine {
	router := gin.Default()

	classesRepo := postgres.NewClassesRepo(db)
	confirmedBookingsRepo := postgres.NewConfirmedBookingsRepo(db)
	pendingBookingsRepo := postgres.NewPendingBookingsRepo(db)
	tokenGenerator := token.NewGenerator()
	emailSender := email.NewSender(
		cfg.EmailSender.Host,
		cfg.EmailSender.Port,
		cfg.EmailSender.User,
		cfg.EmailSender.Password,
		cfg.EmailSender.FromName,
	)

	classesHandler := classes.NewHandler(classesRepo)
	bookHandler := book.NewHandler()
	submitHandler := submit.NewHandler(
		classesRepo,
		confirmedBookingsRepo,
		pendingBookingsRepo,
		tokenGenerator,
		emailSender,
	)
	confirmationHandler := confirmation.NewHandler(confirmedBookingsRepo, pendingBookingsRepo)

	router.Static("/static", "./static")
	router.LoadHTMLGlob("templates/*")
	api := router.Group("/")

	{
		api.GET("/classes", classesHandler.Handle)
	}
	{
		api.POST("/book", bookHandler.Handle)
		api.POST("/submit", submitHandler.Handle)
		api.POST("/confirmation", confirmationHandler.Handle)
	}

	return router
}

func runServer(srv *http.Server, cfg *configuration.Configuration) {
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
