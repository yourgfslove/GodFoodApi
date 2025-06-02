package getRestaurantByID

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/yourgfslove/GodFoodApi/internal/database"
	"github.com/yourgfslove/GodFoodApi/internal/lib/api/response"
	"github.com/yourgfslove/GodFoodApi/internal/lib/logger/sl"
	"log/slog"
	"net/http"
	"strconv"
)

type RestaurantGetter interface {
	GetRestaurantAndMenuByID(ctx context.Context, id int32) ([]database.GetRestaurantAndMenuByIDRow, error)
}

type Response struct {
	response.Response
	RestaurantID      int32  `json:"restaurant_id"`
	RestaurantName    string `json:"restaurant_name"`
	RestaurantAddress string `json:"restaurant_address"`
	RestaurantPhone   string `json:"restaurant_phone"`
	MenuItems         []item `json:"menu_items"`
}

type item struct {
	ItemID          int32   `json:"item_id"`
	ItemName        string  `json:"item_name"`
	ItemPrice       float64 `json:"item_price"`
	ItemDescription string  `json:"item_description"`
}

func New(log *slog.Logger, getter RestaurantGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.restaurants.getRestaurantByID"
		log = log.With(
			slog.String("op", op),
			slog.String("request-id", middleware.GetReqID(r.Context())))
		restaurantID := chi.URLParam(r, "id")
		parsedRestaurantId, err := strconv.ParseInt(restaurantID, 10, 32)
		if err != nil {
			log.Info("invalid restaurant_id", sl.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("Invalid restaurant ID"))
			return
		}
		restaurant, err := getter.GetRestaurantAndMenuByID(r.Context(), int32(parsedRestaurantId))
		if err != nil {
			log.Info("No restaurant by following id", sl.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("Restaurant not found"))
			return
		}
		if len(restaurant) == 0 {
			log.Info("no restaurant by following id", sl.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("Restaurant not found"))
			return
		}
		resp := Response{
			Response:          response.OK(),
			RestaurantID:      restaurant[0].RestaurantID,
			RestaurantName:    restaurant[0].RestaurantName.String,
			RestaurantAddress: restaurant[0].RestaurantAddress.String,
			RestaurantPhone:   restaurant[0].RestaurantPhone,
			MenuItems:         make([]item, 0, len(restaurant)),
		}
		for _, v := range restaurant {
			if v.Available.Bool {
				menuItem := item{
					ItemID:          v.MenuItemID,
					ItemName:        v.MenuItemName,
					ItemPrice:       v.Price,
					ItemDescription: v.Description.String,
				}
				resp.MenuItems = append(resp.MenuItems, menuItem)
			}
		}
		log.Info("Got restaurant")
		render.JSON(w, r, resp)
	}
}
