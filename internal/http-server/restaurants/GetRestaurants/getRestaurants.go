package GetRestaurants

import (
	"context"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/yourgfslove/GodFoodApi/internal/database"
	"github.com/yourgfslove/GodFoodApi/internal/lib/api/response"
	"log/slog"
	"net/http"
)

type restaurantsGetter interface {
	GetUsersByRole(ctx context.Context, userRole string) ([]database.User, error)
}

type Response struct {
	response.Response
	Restaurants []Restaurant `json:"restaurants"`
}

type Restaurant struct {
	Name         string `json:"name"`
	Address      string `json:"address"`
	Phone        string `json:"phone"`
	RestaurantID int32  `json:"restaurant_id"`
}

func New(log *slog.Logger, getter restaurantsGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.restaurants.GetRestaurants"
		log = log.With(slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())))
		restaurants, err := getter.GetUsersByRole(r.Context(), "restaurant")
		if err != nil {
			log.Info("cant't get restaurants")
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("can not get restaurants"))
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
			Response:    response.OK(),
			Restaurants: restaurantList,
		})
	}
}
