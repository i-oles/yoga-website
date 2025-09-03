package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	classesService "main/internal/application/classes"
	"main/internal/application/confirmation"
	"main/internal/application/pending"
	"main/internal/infrastructure/configuration"
	"main/internal/infrastructure/generator/token"
	dbModels "main/internal/infrastructure/models/db/classes"
	"main/internal/infrastructure/models/db/bookings"
	"main/internal/infrastructure/models/db/pendingbookings"
	sqliteRepo "main/internal/infrastructure/repository/sqlite"
	"main/internal/infrastructure/sender/gmail"
	allConfirmedBooking "main/internal/interfaces/http/api/getallbookings"
	"main/internal/interfaces/http/api/createclasses"
	"main/internal/interfaces/http/endpoints/cancel"
	pendingCancel "main/internal/interfaces/http/endpoints/pendingoperation/cancelbooking"
	"main/internal/interfaces/http/err/handler"
	logWrapper "main/internal/interfaces/http/err/wrapper"
	confirmationCancel "main/internal/interfaces/http/html/cancelbooking"
	"main/internal/interfaces/http/html/home"
	"main/internal/interfaces/http/middleware"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		slog.Error("failed to loading configuration", slog.String("err", err.Error()))
		os.Exit(1)
	}

	db, err := gorm.Open(sqlite.Open(cfg.DBPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		panic(err)
	}

	slog.Info("Successfully connected to database")

	err = db.AutoMigrate(
		&dbModels.SQLClass{},
		&pendingbookings.SQLPendingOperation{},
		&bookings.SQLBooking{},
	)
	if err != nil {
		log.Fatal(err)
	}

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

	if cfg.LogConfig {
		slog.Info(cfg.Pretty())
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	return cfg, nil
}

func setupRouter(db *gorm.DB, cfg *configuration.Configuration) *gin.Engine {
	router := gin.Default()

	classesRepo := sqliteRepo.NewClassesRepo(db)
	confirmedBookingsRepo := sqliteRepo.NewBookingsRepo(db)
	pendingOperationsRepo := sqliteRepo.NewPendingOperationsRepo(db)

	tokenGenerator := token.NewGenerator()
	emailSender := gmail.NewSender(
		cfg.EmailSender.Host,
		cfg.EmailSender.Port,
		cfg.EmailSender.User,
		cfg.EmailSender.Password,
		cfg.EmailSender.FromName,
		cfg.ConfirmationCreateEmailTmplPath,
		cfg.ConfirmationCancelEmailTmplPath,
		cfg.ConfirmationFinalEmailTmplPath,
	)

	classesService := classesService.New(classesRepo)
	confirmationService := confirmation.New(classesRepo, confirmedBookingsRepo, pendingOperationsRepo, emailSender)
	pendingOperationsService := pending.New(
		classesRepo,
		pendingOperationsRepo,
		confirmedBookingsRepo,
		tokenGenerator,
		emailSender,
		cfg.DomainAddr,
	)

	var errorHandler handler.IErrorHandler

	errorHandler = handler.NewErrorHandler()
	errorHandler = logWrapper.NewErrorHandler(errorHandler, cfg.LogBusinessErrors)

	homeHandler := home.NewHandler(classesService)
	pendingBookingFormHandler := confirmationCancel.NewHandler()
	cancelHandler := cancel.NewHandler()
	pendingBookingHandler := confirmationCancel.NewHandler(
		pendingOperationsService,
		errorHandler,
	)
	pendingOperationCancelHandler := pendingCancel.NewHandler(
		pendingOperationsService,
		errorHandler,
	)

	createBookingHandler := confirmationCancel.NewHandler(
		confirmationService,
		errorHandler,
	)

	cancelBookingHandler := confirmationCancel.NewHandler(
		confirmationService,
		errorHandler,
	)

	authMiddleware := middleware.Auth(cfg.AuthSecret)

	createClassHandler := createclasses.NewHandler(classesService, errorHandler)
	getAllBookingsHandler := allConfirmedBooking.NewHandler(confirmedBookingsRepo, errorHandler)
	deleteBookingHandler := deleteconfirmedbooking.NewHandler(confirmedBookingsRepo, errorHandler)

	router.Static("web/static", "./web/static")
	router.LoadHTMLGlob("web/templates/*")
	api := router.Group("/")

	// HTML
	{
		api.GET("/", homeHandler.Handle) // home site
		api.POST("/bookings", createBookingHandler.Handle) // creates booking
		api.DELETE("/bookings/:id", cancelBookingHandler.Handle) // deletes booking
		api.POST("/bookings/pending", pendingBookingHandler.Handle) // creates pending booking
		api.GET("/bookings/pending/new", pendingBookingFormHandler.Handle) // opens a form to pending booking
	}
	// API
	{
		api.GET("/api/v1/bookings", authMiddleware, getAllBookingsHandler.Handle) // gets all bookings
		api.DELETE("/api/v1/bookings/:id", authMiddleware, deleteBookingHandler.Handle) // deletes booking
		api.POST("/api/v1/classes", authMiddleware, createClassHandler.Handle) // creates classes
		api.PATCH("/api/v1/classes/:id"), authMiddleware, updateClassHandler.Handle) // updates class
		api.DELETE("/api/v1/classes/:id"), authMiddleware, deleteClassHandler.Handle) // deletes class
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
