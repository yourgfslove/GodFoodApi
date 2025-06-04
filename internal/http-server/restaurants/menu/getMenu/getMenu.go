package getMenu

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/yourgfslove/GodFoodApi/internal/database"
	"github.com/yourgfslove/GodFoodApi/internal/lib/api/response"
	"log/slog"
	"net/http"
	"strconv"
)

type menuGetter interface {
	GetMenu(ctx context.Context, restaurantID int32) ([]database.Menuitem, error)
}

type Response struct {
	RestaurantID int32  `json:"restaurant_id"`
	Menu         []Item `json:"menu"`
}

type Item struct {
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
	Available   bool    `json:"available"`
}

func New(log *slog.Logger, getter menuGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.restaurants.getMenu"
		log = log.With(slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())))
		restaurantID := chi.URLParam(r, "id")
		if restaurantID == "" {
			log.Info("missing restaurant_id")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("missing restaurant_id"))
			return
		}
		IntRestaurantID, err := strconv.ParseInt(restaurantID, 10, 32)
		if err != nil {
			log.Info("invalid restaurant_id")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("restaurant_id is not a number"))
			return
		}
		log.Info("restaurant_id is parsed")
		menu, err := getter.GetMenu(r.Context(), int32(IntRestaurantID))
		if err != nil {
			log.Info("wrong ID")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("No restaurant found"))
			return
		}
		if menu == nil {
			log.Info("restaurant not found")
			w.WriteHeader(http.StatusNotFound)
			render.JSON(w, r, response.Error("restaurant not found"))
			return
		}
		resMenu := make([]Item, 0, len(menu))
		for _, i := range menu {
			resMenu = append(resMenu, Item{
				Name:        i.Name,
				Price:       i.Price,
				Description: i.Description.String,
				Available:   i.Available.Bool,
			})
		}
		render.JSON(w, r, Response{
			RestaurantID: int32(IntRestaurantID),
			Menu:         resMenu,
		})
	}
}
