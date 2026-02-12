package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"main/internal/application/bookings"
	"main/internal/application/classes"
	"main/internal/application/passes"
	"main/internal/application/pendingbookings"
	"main/internal/infrastructure/configuration"
	"main/internal/infrastructure/generator/token"
	dbModels "main/internal/infrastructure/models/db"
	sqliteRepo "main/internal/infrastructure/repository/sqlite"
	"main/internal/infrastructure/sender/gmail"
	apiErrs "main/internal/interfaces/http/api/errs"
	apiErrHandler "main/internal/interfaces/http/api/errs/handler"
	"main/internal/interfaces/http/api/errs/wrapper"
	"main/internal/interfaces/http/api/handlers/activatepass"
	"main/internal/interfaces/http/api/handlers/createclasses"
	"main/internal/interfaces/http/api/handlers/deletebooking"
	"main/internal/interfaces/http/api/handlers/deleteclass"
	"main/internal/interfaces/http/api/handlers/listbookings"
	"main/internal/interfaces/http/api/handlers/listbookingsbyclass"
	"main/internal/interfaces/http/api/handlers/listclasses"
	"main/internal/interfaces/http/api/handlers/listpendingbookings"
	"main/internal/interfaces/http/api/handlers/updateclass"
	viewErrs "main/internal/interfaces/http/html/errs"
	viewErrHandler "main/internal/interfaces/http/html/errs/handler"
	logWrapper "main/internal/interfaces/http/html/errs/wrapper"
	"main/internal/interfaces/http/html/handlers/cancelbooking"
	"main/internal/interfaces/http/html/handlers/cancelbookingform"
	"main/internal/interfaces/http/html/handlers/createbooking"
	"main/internal/interfaces/http/html/handlers/home"
	creatependingbooking "main/internal/interfaces/http/html/handlers/pendingbooking"
	"main/internal/interfaces/http/html/handlers/pendingbookingform"
	"main/internal/interfaces/http/middleware"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"golang.org/x/time/rate"
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
		&dbModels.SQLPendingBooking{},
		&dbModels.SQLBooking{},
		&dbModels.SQLPass{},
	)
	if err != nil {
		log.Fatal(err)
	}

	cleanUpPendingBookingsDBAsync(db)

	router := setupRouter(db, cfg)

	srv := &http.Server{
		Addr:              cfg.ListenAddress,
		Handler:           router,
		ReadHeaderTimeout: cfg.ReadTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
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
	passesRepo := sqliteRepo.NewPassesRepo(db)

	tokenGenerator := token.NewGenerator()
	emailSender := gmail.NewSender(
		cfg.EmailSender.Host,
		cfg.EmailSender.Port,
		cfg.EmailSender.User,
		cfg.EmailSender.Password,
		cfg.EmailSender.FromName,
		cfg.BaseSenderTmplPath,
		cfg.EmailSender.SkipVerification,
	)

	classesService := classes.NewService(classesRepo, bookingsRepo, passesRepo, emailSender)
	bookingsService := bookings.NewService(
		classesRepo, bookingsRepo, pendingBookingsRepo, passesRepo, emailSender, cfg.DomainAddr,
	)
	pendingBookingsService := pendingbookings.NewService(
		classesRepo,
		pendingBookingsRepo,
		bookingsRepo,
		tokenGenerator,
		emailSender,
		cfg.DomainAddr,
	)
	passesService := passes.NewService(passesRepo, bookingsRepo, emailSender)

	var viewErrorHandler viewErrs.IErrorHandler

	viewErrorHandler = viewErrHandler.NewErrorHandler()
	viewErrorHandler = logWrapper.NewErrorHandler(viewErrorHandler, cfg.LogBusinessErrors)

	// HTML
	rateLimiterMiddleware := middleware.GlobalRateLimit

	homeHandler := home.NewHandler(classesService, viewErrorHandler, cfg.IsVacation)
	createBookingHandler := createbooking.NewHandler(bookingsService, viewErrorHandler)
	cancelBookingHandler := cancelbooking.NewHandler(bookingsService, viewErrorHandler)
	createPendingBookingHandler := creatependingbooking.NewHandler(pendingBookingsService, viewErrorHandler)
	pendingBookingFormHandler := pendingbookingform.NewHandler()
	cancelBookingFormHandler := cancelbookingform.NewHandler(bookingsService, viewErrorHandler)
	{
		// home
		api.GET("/", homeHandler.Handle)

		// bookings
		// this endpoint should be POST according to REST, it is GET - confirmation link sent via email
		api.GET("/bookings", createBookingHandler.Handle)
		api.DELETE("/bookings/:id", cancelBookingHandler.Handle)
		api.GET("/bookings/:id/cancel_form", cancelBookingFormHandler.Handle)

		// pending_bookings
		api.GET("/classes/:class_id/pending_bookings/form", pendingBookingFormHandler.Handle)

		requestLimiter := rate.NewLimiter(rate.Limit(1), 2)
		api.POST("/pending_bookings", rateLimiterMiddleware(requestLimiter), createPendingBookingHandler.Handle)
	}

	var apiErrorHandler apiErrs.IErrorHandler

	apiErrorHandler = apiErrHandler.NewErrorHandler()
	apiErrorHandler = wrapper.NewErrorHandler(apiErrorHandler)

	// API
	authMiddleware := middleware.Auth(cfg.AuthSecret)

	createClassHandler := createclasses.NewHandler(classesService, apiErrorHandler)
	getClassesHandler := listclasses.NewHandler(classesService, apiErrorHandler)
	updateClassHandler := updateclass.NewHandler(classesService, apiErrorHandler)
	deleteClassHandler := deleteclass.NewHandler(classesService, apiErrorHandler)
	listBookingsHandler := listbookings.NewHandler(bookingsRepo, apiErrorHandler)
	listBookingsByClassHandler := listbookingsbyclass.NewHandler(bookingsRepo, apiErrorHandler)
	deleteBookingHandler := deletebooking.NewHandler(bookingsService, apiErrorHandler)
	listPendingBookingsHandler := listpendingbookings.NewHandler(pendingBookingsRepo, apiErrorHandler)
	activatePassHandler := activatepass.NewHandler(passesService, apiErrorHandler)

	{
		api.GET("/api/v1/bookings", authMiddleware, listBookingsHandler.Handle)
		api.DELETE("/api/v1/bookings/:booking_id", authMiddleware, deleteBookingHandler.Handle)
		api.GET("api/v1/pending_bookings", authMiddleware, listPendingBookingsHandler.Handle)
		api.POST("/api/v1/classes", authMiddleware, createClassHandler.Handle)
		api.GET("/api/v1/classes", authMiddleware, getClassesHandler.Handle)
		api.PATCH("/api/v1/classes/:class_id", authMiddleware, updateClassHandler.Handle)
		api.DELETE("/api/v1/classes/:class_id", authMiddleware, deleteClassHandler.Handle)
		api.GET("/api/v1/classes/:class_id/bookings", authMiddleware, listBookingsByClassHandler.Handle)
		api.PUT("/api/v1/passes", authMiddleware, activatePassHandler.Handle)
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

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ContextTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", slog.String("err", err.Error()))
	}

	slog.Info("Server stopped")
}

func cleanUpPendingBookingsDBAsync(db *gorm.DB) {
	go func() {
		//nolint
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		oneHourAgo := time.Now().UTC().Add(-1 * time.Hour)

		result := db.WithContext(ctx).
			Where("created_at < ?", oneHourAgo).
			Delete(&dbModels.SQLPendingBooking{})

		if result.Error != nil {
			if errors.Is(result.Error, context.DeadlineExceeded) {
				slog.Warn("cleanup timeout exceeded")
			} else {
				slog.Error("failed to cleanup pending bookings async",
					slog.String("err", result.Error.Error()))
			}

			return
		}

		slog.Info("Cleaned up pending bookings", slog.Int64("rows_deleted", result.RowsAffected))
	}()
}
