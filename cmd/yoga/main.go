package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"main/internal/application/bookings"
	"main/internal/application/classes"
	"main/internal/application/pendingbookings"
	"main/internal/infrastructure/configuration"
	"main/internal/infrastructure/generator/token"
	classesDBModels "main/internal/infrastructure/models/db"
	sqliteRepo "main/internal/infrastructure/repository/sqlite"
	"main/internal/infrastructure/sender/gmail"
	apiErrs "main/internal/interfaces/http/api/errs"
	apiErrHandler "main/internal/interfaces/http/api/errs/handler"
	"main/internal/interfaces/http/api/errs/wrapper"
	"main/internal/interfaces/http/api/handlers/allbookings"
	"main/internal/interfaces/http/api/handlers/allbookingsforclass"
	"main/internal/interfaces/http/api/handlers/createclasses"
	"main/internal/interfaces/http/api/handlers/deleteclass"
	viewErrs "main/internal/interfaces/http/html/errs"
	viewErrHandler "main/internal/interfaces/http/html/errs/handler"
	logWrapper "main/internal/interfaces/http/html/errs/wrapper"
	"main/internal/interfaces/http/html/handlers/cancelbooking"
	"main/internal/interfaces/http/html/handlers/cancelbookingform"
	"main/internal/interfaces/http/html/handlers/createbooking"
	"main/internal/interfaces/http/html/handlers/home"
	"main/internal/interfaces/http/html/handlers/pendingbooking"
	"main/internal/interfaces/http/html/handlers/pendingbookingform"
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
		&classesDBModels.SQLClass{},
		&classesDBModels.SQLPendingBooking{},
		&classesDBModels.SQLBooking{},
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

	router.Static("web/static", "./web/static")
	router.LoadHTMLGlob("web/templates/*")
	api := router.Group("/")

	classesRepo := sqliteRepo.NewClassesRepo(db)
	bookingsRepo := sqliteRepo.NewBookingsRepo(db)
	pendingBookingsRepo := sqliteRepo.NewPendingBookingsRepo(db)

	tokenGenerator := token.NewGenerator()
	emailSender := gmail.NewSender(
		cfg.EmailSender.Host,
		cfg.EmailSender.Port,
		cfg.EmailSender.User,
		cfg.EmailSender.Password,
		cfg.EmailSender.FromName,
		cfg.ConfirmationRequestEmailTmplPath,
		cfg.ConfirmationEmailTmplPath,
	)

	classesService := classes.NewService(classesRepo, bookingsRepo)
	bookingsService := bookings.NewService(classesRepo, bookingsRepo, pendingBookingsRepo, emailSender, cfg.DomainAddr)
	pendingBookingsService := pendingbookings.NewService(
		classesRepo,
		pendingBookingsRepo,
		bookingsRepo,
		tokenGenerator,
		emailSender,
		cfg.DomainAddr,
	)

	var viewErrorHandler viewErrs.IErrorHandler

	viewErrorHandler = viewErrHandler.NewErrorHandler()
	viewErrorHandler = logWrapper.NewErrorHandler(viewErrorHandler, cfg.LogBusinessErrors)

	// HTML
	homeHandler := home.NewHandler(classesService, viewErrorHandler, cfg.IsVacation)
	createBookingHandler := createbooking.NewHandler(bookingsService, viewErrorHandler)
	cancelBookingHandler := cancelbooking.NewHandler(bookingsService, viewErrorHandler)
	pendingBookingHandler := pendingbooking.NewHandler(pendingBookingsService, viewErrorHandler)
	pendingBookingFormHandler := pendingbookingform.NewHandler()
	cancelBookingFormHandler := cancelbookingform.NewHandler(bookingsService, viewErrorHandler)
	{
		api.GET("/", homeHandler.Handle) // home site

		//bookings
		//this endpoint should be POST according to REST, but it is GET - from confirmation link sent via email
		api.GET("/bookings", createBookingHandler.Handle)                     // creates booking
		api.DELETE("/bookings/:id", cancelBookingHandler.Handle)              // deletes booking
		api.GET("/bookings/:id/cancel_form", cancelBookingFormHandler.Handle) // renders cancel booking form

		//pending_bookings
		api.GET("/classes/:class_id/pending_bookings/form", pendingBookingFormHandler.Handle) // renders a form to pending booking
		api.POST("/pending_bookings", pendingBookingHandler.Handle)                           // creates pending booking
	}

	var apiErrorHandler apiErrs.IErrorHandler

	apiErrorHandler = apiErrHandler.NewErrorHandler()
	apiErrorHandler = wrapper.NewErrorHandler(apiErrorHandler)

	// API
	authMiddleware := middleware.Auth(cfg.AuthSecret)
	createClassHandler := createclasses.NewHandler(classesService, apiErrorHandler)
	//updateClassHandler := updateclass.NewHandler(classesService, apiErrorHandler)
	deleteClassHandler := deleteclass.NewHandler(classesService, apiErrorHandler)
	getAllBookingsHandler := allbookings.NewHandler(bookingsRepo, apiErrorHandler)
	getAllBookingsForClassHandler := allbookingsforclass.NewHandler(bookingsRepo, apiErrorHandler)
	//deleteBookingHandler := deletebooking.NewHandler(bookingsRepo, apiErrorHandler)
	//getAllPendingBookingsHandler := allpendingbookings.NewHandler(bookingsRepo, apiErrorHandler)

	{
		api.GET("/api/v1/bookings", authMiddleware, getAllBookingsHandler.Handle) // gets all bookings
		//api.DELETE("/api/v1/bookings/:id", authMiddleware, deleteBookingHandler.Handle)         // deletes booking
		//api.GET("api/v1/bookings/pending", authMiddleware, getAllPendingBookingsHandler.Handle) //gets all
		api.POST("/api/v1/classes", authMiddleware, createClassHandler.Handle) // creates classes
		//api.PATCH("/api/v1/classes/:id"), authMiddleware, updateClassHandler.Handle)            // updates class
		api.DELETE("/api/v1/classes/:class_id", authMiddleware, deleteClassHandler.Handle)           // deletes class
		api.GET("/api/v1/classes/:class_id/bookings", authMiddleware, getAllBookingsForClassHandler.Handle)
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
