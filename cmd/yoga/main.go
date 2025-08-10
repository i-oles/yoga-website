package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"main/configuration"
	"main/internal/api/handlers/book"
	"main/internal/api/handlers/classes"
	confirmationCancel "main/internal/api/handlers/confirmation/cancel"
	confirmationCreate "main/internal/api/handlers/confirmation/create"
	pendingCancel "main/internal/api/handlers/pending/cancel"
	pendingCreate "main/internal/api/handlers/pending/create"
	classesService "main/internal/domain/services/classes"
	"main/internal/domain/services/confirmation"
	"main/internal/domain/services/pending"
	"main/internal/errs"
	"main/internal/errs/app"
	log2 "main/internal/errs/log"
	"main/internal/generator/token"
	"main/internal/infrastructure/repositories/postgres"
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
	pendingOperationsRepo := postgres.NewPendingOperationsRepo(db)

	tokenGenerator := token.NewGenerator()
	emailSender := email.NewSender(
		cfg.EmailSender.Host,
		cfg.EmailSender.Port,
		cfg.EmailSender.User,
		cfg.EmailSender.Password,
		cfg.EmailSender.FromName,
	)

	classesService := classesService.New(classesRepo)
	confirmationService := confirmation.New(classesRepo, confirmedBookingsRepo, pendingOperationsRepo)
	pendingOperationsService := pending.New(
		classesRepo,
		pendingOperationsRepo,
		tokenGenerator,
		emailSender,
		cfg.DomainAddr,
	)

	var errorHandler errs.ErrorHandler

	errorHandler = app.NewErrorHandler()
	if cfg.LogErrors {
		errorHandler = log2.NewErrorHandler(errorHandler)
	}

	classesHandler := classes.NewHandler(classesService)
	bookHandler := book.NewHandler()
	pendingCreateHandler := pendingCreate.NewHandler(
		pendingOperationsService,
		errorHandler,
	)
	pendingCancelHandler := pendingCancel.NewHandler(
		pendingOperationsService,
		errorHandler,
	)

	confirmationCreateHandler := confirmationCreate.NewHandler(
		confirmationService,
		errorHandler,
	)

	confirmationCancelHandler := confirmationCancel.NewHandler(
		confirmationService,
		errorHandler,
	)

	router.Static("/static", "./static")
	router.LoadHTMLGlob("templates/*")
	api := router.Group("/")

	{
		api.GET("/classes", classesHandler.Handle)
		api.GET("/confirmation/create_booking", confirmationCreateHandler.Handle)
		api.GET("/confirmation/cancel_booking", confirmationCancelHandler.Handle)
	}
	{
		api.POST("/book", bookHandler.Handle)
		api.POST("/pending_operation/:class_id/create_booking", pendingCreateHandler.Handle)
		api.POST("/pending_operation/:class_id/cancel_booking", pendingCancelHandler.Handle)

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
