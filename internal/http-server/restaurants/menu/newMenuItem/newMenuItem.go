package newMenuItem

import (
	"context"
	"database/sql"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/yourgfslove/GodFoodApi/internal/database"
	"github.com/yourgfslove/GodFoodApi/internal/lib/api/response"
	"github.com/yourgfslove/GodFoodApi/internal/lib/auth/JWT"
	"github.com/yourgfslove/GodFoodApi/internal/lib/auth/getToken"
	"github.com/yourgfslove/GodFoodApi/internal/lib/logger/sl"
	"log/slog"
	"net/http"
)

type Request struct {
	Price       float64 `json:"price"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Available   bool    `json:"available"`
}

type Response struct {
	response.Response
	ID           int32   `json:"id"`
	RestaurantID int32   `json:"restaurant_id"`
	Name         string  `json:"name"`
	Price        float64 `json:"price"`
	Description  string  `json:"description,omitempty"`
	Available    bool    `json:"available"`
}

type menuItemCreater interface {
	CreateMenuItem(ctx context.Context, arg database.CreateMenuItemParams) (database.Menuitem, error)
}
type userGetter interface {
	GetUserByID(ctx context.Context, id int32) (database.User, error)
}

func New(log *slog.Logger, creater menuItemCreater, userGetter userGetter, tokenSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.restaurants.newMenuItem"
		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())))
		var req Request
		if err := render.DecodeJSON(r.Body, &req); err != nil {
			log.Info("failed to decode JSON", sl.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("failed to decode JSON"))
			return
		}
		token, err := getToken.GetTokenFromHeader(r.Header, "Bearer")
		if err != nil {
			log.Info("failed to get token", sl.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("failed to get token"))
			return
		}
		userID, err := JWT.ValidateJWT(token, tokenSecret)
		if err != nil {
			log.Info("failed to validate token", sl.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("failed to validate token"))
			return
		}
		userRest, err := userGetter.GetUserByID(r.Context(), userID)
		if err != nil {
			log.Info("wrong ID")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("wrong ID"))
			return
		}
		if userRest.UserRole != "restaurant" {
			log.Info("wrong role")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("wrong role"))
			return
		}
		newItem, err := creater.CreateMenuItem(context.Background(), database.CreateMenuItemParams{
			RestaurantID: userID,
			Name:         req.Name,
			Price:        req.Price,
			Description:  sql.NullString{String: req.Description, Valid: req.Description != ""},
			Available:    sql.NullBool{Bool: req.Available, Valid: true},
		})
		if err != nil {
			log.Info("failed to create menu item", sl.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("failed to create menu item"))
			return
		}
		render.JSON(w, r, Response{
			Response:     response.OK(),
			ID:           newItem.ID,
			RestaurantID: newItem.RestaurantID,
			Name:         newItem.Name,
			Price:        newItem.Price,
			Description:  newItem.Description.String,
			Available:    newItem.Available.Bool,
		})
	}
}
