package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"main/configuration"
	"main/internal/handler/book"
	"main/internal/handler/classes"
	"main/internal/repository/postgres"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		slog.Error("failed to loading configuration", slog.String("err", err.Error()))
		os.Exit(1)
	}

	connStr := fmt.Sprintf("dbname=%s user=%s password=%s host=localhost sslmode=disable",
		cfg.PostgresDBName,
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
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

	router := setupRouter(db)

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
	err := godotenv.Load()
	if err != nil {
		slog.Info("No .env file found, using environment variables...")
	}

	postgresUser := os.Getenv("POSTGRES_USER")
	postgresPassword := os.Getenv("POSTGRES_PASSWORD")
	if postgresUser == "" || postgresPassword == "" {
		return configuration.Configuration{},
			errors.New("provide envs for postgres access")
	}

	var cfg configuration.Configuration

	err = configuration.GetConfig("./config", &cfg)
	if err != nil {
		return configuration.Configuration{},
			fmt.Errorf("error loading configuration: %w", err)
	}

	slog.Info(cfg.Pretty())

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	return cfg, nil
}

func setupRouter(db *sql.DB) *gin.Engine {
	router := gin.Default()

	classesRepo := postgres.NewClassesRepo(db)
	classesHandler := classes.NewHandler(classesRepo)

	bookHandler := book.NewHandler()

	router.Static("/static", "./static")
	router.LoadHTMLGlob("templates/*")
	api := router.Group("/")

	{
		api.GET("/classes", classesHandler.Handle)
	}

	{
		api.POST("/book", bookHandler.Handle)
	}

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
