package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/swaggo/http-swagger"
	_ "github.com/yourgfslove/GodFoodApi/docs"
	"github.com/yourgfslove/GodFoodApi/internal/config"
	"github.com/yourgfslove/GodFoodApi/internal/database"
	myrouter "github.com/yourgfslove/GodFoodApi/internal/http-server/router"
	"github.com/yourgfslove/GodFoodApi/internal/lib/logger/sl"
	"log/slog"
	"net/http"
	"os"
)

// @title GodFood API
// @version 1.0
// @description REST API for food delivery
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Введите токен в формате: Bearer {token}
func main() {
	fmt.Println("Starting server...")
	cfg := config.MustLoadConfig()
	log := setupLogger(cfg.Env)

	log.Info("starting server")
	log.Debug("Debug logging enabled")
	storage, err := sql.Open("postgres", cfg.StorageURL)
	if err != nil {
		log.Error("failed init storage", err)
		os.Exit(1)
	}
	DBQueries := database.New(storage)
	router := myrouter.New(log)
	deps := &myrouter.Deps{
		Storage: DBQueries,
		Logger:  log,
		Cfg: struct {
			SecretJWT string
		}{},
	}
	router.Get("/docs/*", httpSwagger.WrapHandler)
	myrouter.SetupRoutes(router, deps)
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
