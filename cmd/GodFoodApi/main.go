package main

import (
	"database/sql"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"
	"github.com/yourgfslove/GodFoodApi/internal/config"
	"github.com/yourgfslove/GodFoodApi/internal/database"
	"github.com/yourgfslove/GodFoodApi/internal/http-server/auth/login"
	"github.com/yourgfslove/GodFoodApi/internal/http-server/auth/register"
	mwLogger "github.com/yourgfslove/GodFoodApi/internal/http-server/middleware/logger"
	"github.com/yourgfslove/GodFoodApi/internal/http-server/orders/getOrderByID"
	"github.com/yourgfslove/GodFoodApi/internal/http-server/orders/getOrdersForUser"
	"github.com/yourgfslove/GodFoodApi/internal/http-server/orders/placeorder"
	"github.com/yourgfslove/GodFoodApi/internal/http-server/restaurants/GetRestaurants"
	"github.com/yourgfslove/GodFoodApi/internal/http-server/restaurants/getRestaurantByID"
	"github.com/yourgfslove/GodFoodApi/internal/http-server/restaurants/menu/getMenu"
	"github.com/yourgfslove/GodFoodApi/internal/http-server/restaurants/menu/newMenuItem"
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
		log.Error("failed init storage", err)
		os.Exit(1)
	}
	DBQueries := database.New(storage)
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	router.Post("/register", register.New(log, DBQueries, DBQueries, cfg.SecretJWT))
	router.Get("/login", login.New(log, DBQueries, DBQueries, cfg.SecretJWT))
	router.Post("/restaurants/menuItems", newMenuItem.New(log, DBQueries, DBQueries, cfg.SecretJWT))
	router.Get("/restaurants/{id}/menuItems", getMenu.New(log, DBQueries))
	router.Get("/restaurants", GetRestaurants.New(log, DBQueries))
	router.Post("/orders", placeorder.New(log, DBQueries, DBQueries, DBQueries, cfg.SecretJWT, DBQueries))
	router.Get("/orders", getOrdersForUser.New(log, DBQueries, cfg.SecretJWT))
	router.Get("/restaurants/{id}", getRestaurantByID.New(log, DBQueries))
	router.Get("/orders/{id}", getOrderByID.New(log, DBQueries, DBQueries, cfg.SecretJWT))
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
