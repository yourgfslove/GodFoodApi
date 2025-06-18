package getRestaurantByID

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

type RestaurantGetter interface {
	GetRestaurantAndMenuByID(ctx context.Context, id int32) ([]database.GetRestaurantAndMenuByIDRow, error)
}

type Response struct {
	RestaurantID      int32  `json:"restaurant_id" example:"14"`
	RestaurantName    string `json:"restaurant_name" example:"mac"`
	RestaurantAddress string `json:"restaurant_address" example:"112 address"`
	RestaurantPhone   string `json:"restaurant_phone" example:"89053435656"`
	MenuItems         []item `json:"menu_items"`
}

type item struct {
	ItemID          int32   `json:"item_id" example:"1"`
	ItemName        string  `json:"item_name" example:"cheeseburger"`
	ItemPrice       float64 `json:"item_price" example:"122.00"`
	ItemDescription string  `json:"item_description" example:"burger with cheese"`
}

// Restaurants godoc
// @Summary Получение Ресторана по айди
// @Description Возвращает полную информацию по ресторану(Айди, имя, адрес, телефон) и меню по айди
// @Tags Restaurants
// @Accept json
// @Produce json
// @Param id path int true "ID ресторана"
// @Success 200 {object} getRestaurantByID.Response "Ресторан успешно получен"
// @Failure 400 {object} response.Response "Некорректные данные"
// @Failure 404 {object} response.Response "Ресторан не найден"
// @Router /restaurants/{id} [get]
func New(log *slog.Logger, getter RestaurantGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.restaurants.getRestaurantByID"
		log = log.With(
			slog.String("op", op),
			slog.String("request-id", middleware.GetReqID(r.Context())))
		restaurantID := chi.URLParam(r, "id")

		if restaurantID == "" {
			response.Error(log, w, r, "No ID in URL", "empty ID", http.StatusBadRequest)
			return
		}

		parsedRestaurantId, err := strconv.ParseInt(restaurantID, 10, 32)
		if err != nil {
			response.Error(log, w, r, "Invalid restaurant ID", "Failed to parse ID", http.StatusBadRequest)
			return
		}

		if parsedRestaurantId < 1 {
			response.Error(log, w, r, "invalid restaurant ID", "ID < 1", http.StatusBadRequest)
			return
		}

		restaurant, err := getter.GetRestaurantAndMenuByID(r.Context(), int32(parsedRestaurantId))
		if err != nil {
			response.Error(log, w, r, "Not Found", "no restaurant by folowing ID", http.StatusNotFound)
			return
		}

		if len(restaurant) == 0 {
			response.Error(log, w, r, "Not Found", "empty list", http.StatusNotFound)
			return
		}

		resp := Response{
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
		render.Status(r, http.StatusOK)
		render.JSON(w, r, resp)
	}
}
