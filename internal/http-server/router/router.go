package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	mwLogger "github.com/yourgfslove/GodFoodApi/internal/http-server/middleware/logger"
	"log/slog"
)

func New(log *slog.Logger) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)
	r.Use(middleware.Heartbeat("/ping"))
	r.Use(mwLogger.New(log))
	return r
}
