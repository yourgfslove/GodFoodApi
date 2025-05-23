package main

import (
	"database/sql"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/yourgfslove/GodFoodApi/internal/config"
	"github.com/yourgfslove/GodFoodApi/internal/database"
	"github.com/yourgfslove/GodFoodApi/internal/http-server/auth/register"
	mwLogger "github.com/yourgfslove/GodFoodApi/internal/http-server/middleware/logger"
	"github.com/yourgfslove/GodFoodApi/internal/lib/logger/sl"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	fmt.Println("Starting server...")
	cfg := config.MustLoadConfig()
	log := setupLogger(cfg.Env)

	log.Info("starting server")
	log.Debug("Debug logging enabled")
	storage, err := sql.Open("postgres", cfg.StorageURL)
	if err != nil {
		log.Error("failed init storage")
		os.Exit(1)
	}
	DBQueries := database.New(storage)

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	router.Post("/register", register.New(log, DBQueries))

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
		IdleTimeout:  cfg.IdleTimeout,
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed start server", sl.Err(err))
	}

	log.Error("server shutdown")

}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case "local":
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case "dev":
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case "prod":
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	return log

}
