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
		restaurants, err := getter.GetUsersByRole(context.Background(), "restaurant")
		if err != nil {
			log.Info("cant't get restaurants")
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("can not get restaurants"))
			return
		}
		for _, i := range restaurants {
			render.JSON(w, r, Response{
				Response:     response.OK(),
				Name:         i.UserName.String,
				Address:      i.Address.String,
				Phone:        i.Phone,
				RestaurantID: i.ID,
			})
		}
	}
}
