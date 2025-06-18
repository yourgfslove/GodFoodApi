package getOrderByID

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
	"time"
)

type orderGetter interface {
	GetFullOrderByID(ctx context.Context, id int32) ([]database.GetFullOrderByIDRow, error)
}

type Response struct {
	RestaurantName    string  `json:"restaurant_Name" example:"Mac"`
	RestaurantAddress string  `json:"restaurant_Address" example:"123 address"`
	RestaurantPhone   string  `json:"restaurant_Phone" example:"89055463333"`
	DeliveryAddress   string  `json:"delivery_Address" example:"1222 address"`
	CourierName       string  `json:"courierName" example:"John"`
	UserName          string  `json:"user_name" example:"Bill"`
	Status            string  `json:"status" example:"pending"`
	CreatedAt         string  `json:"created_at" example:"2020-09-20T14:14:15+09:00"`
	Items             []item  `json:"items"`
	TotalPrice        float64 `json:"total_price" example:"300.0"`
}
type item struct {
	ItemName  string  `json:"item_name" example:"burger"`
	ItemPrice float64 `json:"price" example:"60.0"`
	Quantity  int32   `json:"quantity" example:"5"`
}

// orders godoc
// @Summary Получение заказа по айди
// @Description Возвращает полную информацию по заказу(если авторизованный пользователь им владеет)
// @Tags Orders
// @Accept json
// @Produce json
// @Param id path int true "ID Заказа"
// @Success 200 {object} getOrderByID.Response "Заказ успешно получен"
// @Failure 400 {object} response.Response "Неккоректные данные"
// @Failure 401 {object} response.Response "Не авторизован"
// @Failure 403 {object} response.Response "Доступ Запрещен"
// @Failure 500 {object} response.Response "Ошибка сервера"
// @Router /orders/{id} [get]
// @Security BearerAuth
func New(log *slog.Logger, getter orderGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http_server.orders.getOrderByID.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())))

		userID := r.Context().Value("userID").(int32)

		orderID := chi.URLParam(r, "id")
		if orderID == "" {
			response.Error(log, w, r, "No ID in url", "empty ID", http.StatusBadRequest)
			return
		}
		parsedOrderID, err := strconv.ParseInt(orderID, 10, 32)
		if err != nil {
			response.Error(log, w, r, "invalid user ID", "failed to parse userID", http.StatusBadRequest)
			return
		}

		order, err := getter.GetFullOrderByID(r.Context(), int32(parsedOrderID))
		if err != nil {
			response.Error(log, w, r, "Not Found", "No user by following id", http.StatusNotFound)
			return
		}

		if len(order) == 0 {
			response.Error(log, w, r, "Not Found", "No user by following id", http.StatusNotFound)
			return
		}

		if order[0].CustomerID != userID {
			response.Error(log, w, r, "Access denied", "not matching ids", http.StatusForbidden)
			return
		}

		resp := Response{
			RestaurantName:    order[0].RestaurantName.String,
			RestaurantAddress: order[0].RestaurantAddress.String,
			RestaurantPhone:   order[0].RestaurantPhone,
			DeliveryAddress:   order[0].DeliveryAddress,
			CourierName:       order[0].CourierName.String,
			UserName:          order[0].CostomerName.String,
			Status:            order[0].Status,
			CreatedAt:         order[0].CreatedAt.Time.Format(time.RFC3339),
			Items:             []item{},
			TotalPrice:        0,
		}
		for _, v := range order {
			resp.Items = append(resp.Items, item{
				ItemName:  v.MenuItemName,
				ItemPrice: v.Price,
				Quantity:  v.Quanity,
			})
			resp.TotalPrice += v.Price * float64(v.Quanity)
		}

		log.Info("got order")
		render.Status(r, http.StatusOK)
		render.JSON(w, r, resp)
	}
}
