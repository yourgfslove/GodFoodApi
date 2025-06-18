package newMenuItem

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/yourgfslove/GodFoodApi/internal/database"
	"github.com/yourgfslove/GodFoodApi/internal/lib/api/response"
	"github.com/yourgfslove/GodFoodApi/internal/lib/logger/sl"
	"log/slog"
	"net/http"
)

type Request struct {
	Price       float64 `json:"price" example:"52.2"`
	Name        string  `json:"name" example:"Burger"`
	Description string  `json:"description,omitempty" example:"burger with beef"`
	Available   bool    `json:"available"`
}

type Response struct {
	ID           int32   `json:"id" example:"1"`
	RestaurantID int32   `json:"restaurant_id" example:"1"`
	Name         string  `json:"name" example:"Burger"`
	Price        float64 `json:"price" example:"52.2"`
	Description  string  `json:"description,omitempty" example:"burger with beef"`
	Available    bool    `json:"available"`
}

type menuItemCreater interface {
	CreateMenuItem(ctx context.Context, arg database.CreateMenuItemParams) (database.Menuitem, error)
}
type userGetter interface {
	GetUserByID(ctx context.Context, id int32) (database.User, error)
}

// Retaurants godoc
// @Summary Добавление новой позиции в меню
// @Description Создает новую позицию в меню рессторана по JWT
// @Tags Restaurants
// @Accept json
// @Produce json
// @Param request body newMenuItem.Request true "Данные для добавления"
// @Success 200 {object} newMenuItem.Response "Новая позиция успешно добавлена"
// @Failure 400 {object} response.Response "Некорректные данные"
// @Failure 500 {object} response.Response "Серверная Ошибка"
// @Router /restaurants/menuItems [post]
// @Security BearerAuth
func New(log *slog.Logger, creater menuItemCreater, userGetter userGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.restaurants.newMenuItem"
		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())))

		userID, ok := r.Context().Value("userID").(int32)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			render.JSON(w, r, "Not authorized")
			return
		}

		var req Request
		if err := render.DecodeJSON(r.Body, &req); err != nil {
			response.Error(log, w, r,
				"Something went wrong",
				"failed to decode JSON",
				http.StatusInternalServerError)
			return
		}

		userRest, err := userGetter.GetUserByID(r.Context(), userID)
		if err != nil {
			response.Error(log, w, r, "Wrong JWT", "No users for following ID", http.StatusUnauthorized)
			return
		}

		if userRest.UserRole != "restaurant" {
			response.Error(log, w, r, "access denied", "Wrong role", http.StatusForbidden)
			return
		}

		newItem, err := creater.CreateMenuItem(r.Context(), database.CreateMenuItemParams{
			RestaurantID: userID,
			Name:         req.Name,
			Price:        req.Price,
			Description:  sql.NullString{String: req.Description, Valid: req.Description != ""},
			Available:    sql.NullBool{Bool: req.Available, Valid: true},
		})

		if err != nil {
			response.Error(log, w, r,
				"failed to create",
				fmt.Sprintf("failed to create menu item: %v", sl.Err(err)),
				http.StatusInternalServerError)
			return
		}

		render.Status(r, http.StatusCreated)
		render.JSON(w, r, Response{
			ID:           newItem.ID,
			RestaurantID: newItem.RestaurantID,
			Name:         newItem.Name,
			Price:        newItem.Price,
			Description:  newItem.Description.String,
			Available:    newItem.Available.Bool,
		})
	}
}
