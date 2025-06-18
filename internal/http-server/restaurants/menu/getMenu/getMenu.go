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
	RestaurantID int32  `json:"restaurant_id" example:"1"`
	Menu         []Item `json:"menu"`
}

type Item struct {
	Name        string  `json:"name" example:"Cheeseburger"`
	Price       float64 `json:"price" example:"122.00"`
	Description string  `json:"description" example:"burger with cheese"`
	Available   bool    `json:"available"`
}

// Restaurants godoc
// @Summary Получение меню по айди
// @Description Возвращает меню ресторана по айди
// @Tags Restaurants
// @Accept json
// @Produce json
// @Param id path int true "ID ресторана"
// @Success 200 {object} getMenu.Response "Меню успешно получено"
// @Failure 400 {object} response.Response "Некорректные данные"
// @Failure 404 {object} response.Response "Ресторан не найден"
// @Router /restaurants/{id}/menuItems [get]
func New(log *slog.Logger, getter menuGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.restaurants.getMenu"
		log = log.With(slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())))

		restaurantID := chi.URLParam(r, "id")
		if restaurantID == "" {
			response.Error(log, w, r, "No ID in URL", "empty id in URL", http.StatusBadRequest)
			return
		}

		IntRestaurantID, err := strconv.ParseInt(restaurantID, 10, 32)
		if err != nil {
			response.Error(log, w, r, "Invalid ID", "can not parse ID", http.StatusBadRequest)
			return
		}
		log.Info("restaurant_id is parsed")

		menu, err := getter.GetMenu(r.Context(), int32(IntRestaurantID))
		if err != nil {
			response.Error(log, w, r, "Not Found", "no menu found", http.StatusNotFound)
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

		render.Status(r, http.StatusOK)
		render.JSON(w, r, Response{
			RestaurantID: int32(IntRestaurantID),
			Menu:         resMenu,
		})
	}
}
