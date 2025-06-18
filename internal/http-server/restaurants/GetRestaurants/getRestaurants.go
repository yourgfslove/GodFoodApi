package GetRestaurants

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/yourgfslove/GodFoodApi/internal/database"
	"github.com/yourgfslove/GodFoodApi/internal/lib/api/response"
	"github.com/yourgfslove/GodFoodApi/internal/lib/logger/sl"
	"log/slog"
	"net/http"
)

type restaurantsGetter interface {
	GetUsersByRole(ctx context.Context, userRole string) ([]database.User, error)
}

type Response struct {
	Restaurants []Restaurant `json:"restaurants"`
}

type Restaurant struct {
	Name         string `json:"name" example:"Mac"`
	Address      string `json:"address" example:"123 street 1"`
	Phone        string `json:"phone" example:"89055463333"`
	RestaurantID int32  `json:"restaurant_id" example:"1"`
}

// Restaurants godoc
// @Summary Получение всех Ресторанов
// @Description Возвращает полную информацию по всем ресторанам(Айди, имя, адрес, телефон)
// @Tags Restaurants
// @Accept json
// @Produce json
// @Success 200 {object} GetRestaurants.Response "Рестораны успешно получены"
// @Failure 500 {object} response.Response "Ошибка сервера"
// @Router /restaurants [get]
func New(log *slog.Logger, getter restaurantsGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.restaurants.GetRestaurants"
		log = log.With(slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())))

		restaurants, err := getter.GetUsersByRole(r.Context(), "restaurant")
		if err != nil {
			response.Error(log, w, r,
				"can not get restaurants",
				fmt.Sprintf("db err: %v", sl.Err(err)),
				http.StatusInternalServerError)
			return
		}

		restaurantList := make([]Restaurant, 0, len(restaurants))
		for _, i := range restaurants {
			restaurantList = append(restaurantList, Restaurant{
				Name:         i.UserName.String,
				Address:      i.Address.String,
				Phone:        i.Phone,
				RestaurantID: i.ID,
			})
		}
		render.JSON(w, r, Response{
			Restaurants: restaurantList,
		})
	}
}
