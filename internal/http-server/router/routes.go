package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/yourgfslove/GodFoodApi/internal/database"
	"github.com/yourgfslove/GodFoodApi/internal/http-server/auth/login"
	"github.com/yourgfslove/GodFoodApi/internal/http-server/auth/register"
	"github.com/yourgfslove/GodFoodApi/internal/http-server/middleware/middlewareJWT"
	"github.com/yourgfslove/GodFoodApi/internal/http-server/orders/getOrderByID"
	"github.com/yourgfslove/GodFoodApi/internal/http-server/orders/getOrdersForUser"
	"github.com/yourgfslove/GodFoodApi/internal/http-server/orders/getPendingOrders"
	"github.com/yourgfslove/GodFoodApi/internal/http-server/orders/placeorder"
	"github.com/yourgfslove/GodFoodApi/internal/http-server/restaurants/GetRestaurants"
	"github.com/yourgfslove/GodFoodApi/internal/http-server/restaurants/getRestaurantByID"
	"github.com/yourgfslove/GodFoodApi/internal/http-server/restaurants/menu/getMenu"
	"github.com/yourgfslove/GodFoodApi/internal/http-server/restaurants/menu/newMenuItem"
	"log/slog"
)

type Deps struct {
	Storage *database.Queries
	Logger  *slog.Logger
	Cfg     struct {
		SecretJWT string
	}
}

func SetupRoutes(r *chi.Mux, deps *Deps) {
	r.Post("/register", register.New(deps.Logger, deps.Storage, deps.Storage, deps.Cfg.SecretJWT))
	r.Get("/login", login.New(deps.Logger, deps.Storage, deps.Storage, deps.Cfg.SecretJWT))
	r.With(middlewareJWT.AuthJWTMiddleware(deps.Cfg.SecretJWT)).
		Post("/restaurants/menuItems", newMenuItem.New(deps.Logger, deps.Storage, deps.Storage))
	r.Get("/restaurants/{id}/menuItems", getMenu.New(deps.Logger, deps.Storage))
	r.Get("/restaurants", GetRestaurants.New(deps.Logger, deps.Storage))
	r.Get("/restaurants/{id}", getRestaurantByID.New(deps.Logger, deps.Storage))
	r.With(middlewareJWT.AuthJWTMiddleware(deps.Cfg.SecretJWT)).
		Post("/orders", placeorder.New(deps.Logger, deps.Storage, deps.Storage, deps.Storage, deps.Storage))
	r.With(middlewareJWT.AuthJWTMiddleware(deps.Cfg.SecretJWT)).
		Get("/orders", getOrdersForUser.New(deps.Logger, deps.Storage))
	r.With(middlewareJWT.AuthJWTMiddleware(deps.Cfg.SecretJWT)).
		Get("/orders/{id}", getOrderByID.New(deps.Logger, deps.Storage))
	r.With(middlewareJWT.AuthJWTMiddleware(deps.Cfg.SecretJWT)).
		Get("/orders/pending", getPendingOrders.New(deps.Logger, deps.Storage, deps.Storage))
	// router.Patch("/orders/{id}/assign")
	// router.Get("/orders/current")
}
