package getCurrentOrder

import (
	"context"
	"database/sql"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/yourgfslove/GodFoodApi/internal/database"
	"github.com/yourgfslove/GodFoodApi/internal/lib/api/response"
	"log/slog"
	"net/http"
	"time"
)

type curretnOrderGetter interface {
	GetCurrentOrderForCourier(ctx context.Context, courierid sql.NullInt32) ([]database.GetCurrentOrderForCourierRow, error)
}

type courierGetter interface {
	GetUserByID(ctx context.Context, id int32) (database.User, error)
}

type Response struct {
	RestaurantName    string  `json:"restaurant_Name" example:"Mac"`
	RestaurantAddress string  `json:"restaurant_Address" example:"123 address"`
	RestaurantPhone   string  `json:"restaurant_Phone" example:"89056663333"`
	DeliveryAddress   string  `json:"delivery_Address" example:"122 address"`
	Status            string  `json:"status" example:"pending"`
	CreatedAt         string  `json:"created_at" example:"2020-01-01 01:02:03 UTC"`
	Items             []item  `json:"items"`
	Reward            float64 `json:"reward" example:"12.00"`
}
type item struct {
	ItemName string `json:"item_name" example:"Burger with cheese"`
	Quantity int32  `json:"quantity" example:"3"`
}

// Orders godoc
// @Summary Получение нынешнего заказа курьера
// @Description Возвращает полную информацию по заказу, который везет авторизованный курьер
// @Tags Orders
// @Accept json
// @Produce json
// @Success 200 {object} getCurrentOrder.Response "Заказ успешно получен"
// @Failure 403 {object} response.Response "Неавторизован"
// @Failure 404 {object} response.Response "Заказ не найден"
// @Failure 500 {object} response.Response "Ошибка сервера"
// @Router /orders/current [get]
// @Security BearerAuth
func New(log *slog.Logger, getterOrder curretnOrderGetter, getterCourier courierGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.orders.courier.getCurrentOrder"
		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())))

		userID := r.Context().Value("userID").(int32)

		courierInfo, err := getterCourier.GetUserByID(r.Context(), userID)
		if err != nil {
			response.Error(log, w, r, "wrong courier ID", "cannot get courier", http.StatusInternalServerError)
			return
		}

		if courierInfo.UserRole != "courier" {
			response.Error(log, w, r, "Access denied", "wrong role", http.StatusForbidden)
			return
		}

		order, err := getterOrder.GetCurrentOrderForCourier(r.Context(), sql.NullInt32{
			Int32: courierInfo.ID,
			Valid: true})

		if err != nil {
			response.Error(log, w, r, "No current order", "no order", http.StatusNotFound)
			return
		}

		if len(order) == 0 {
			response.Error(log, w, r, "No current order", "no order", http.StatusNotFound)
			return
		}

		resp := Response{
			RestaurantName:    order[0].RestaurantName.String,
			RestaurantAddress: order[0].RestaurantAddress.String,
			RestaurantPhone:   order[0].RestaurantPhone,
			DeliveryAddress:   order[0].DeliveryAddress,
			Status:            order[0].Status,
			CreatedAt:         order[0].CreatedAt.Time.Format(time.RFC3339),
			Items:             []item{},
			Reward:            0,
		}
		for _, v := range order {
			resp.Items = append(resp.Items, item{
				ItemName: v.MenuItemName,
				Quantity: v.Quanity,
			})
			resp.Reward += v.Price * float64(v.Quanity) * 0.05 // could add better version of reward
		}

		log.Info("got order")
		render.Status(r, http.StatusOK)
		render.JSON(w, r, resp)
	}
}
