package getOrdersForUser

import (
	"context"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/yourgfslove/GodFoodApi/internal/database"
	"github.com/yourgfslove/GodFoodApi/internal/lib/api/ordersStruct"
	"github.com/yourgfslove/GodFoodApi/internal/lib/api/response"
	"log/slog"
	"net/http"
)

type OrderGetter interface {
	GetFullOrdersByUserID(ctx context.Context, customerid int32) ([]database.GetFullOrdersByUserIDRow, error)
}

type Response struct {
	Orders []ordersStruct.Order `json:"ordersStruct"`
}

// orders godoc
// @Summary Получение заказов по JWT
// @Description Возвращает полную информацию по заказам для авторизованного пользователя
// @Tags Orders
// @Accept json
// @Produce json
// @Success 200 {object} getOrdersForUser.Response "Заказ успешно получен"
// @Success 204 "Нет заказов"
// @Failure 400 {object} response.Response "Неккоректные данные"
// @Failure 401 {object} response.Response "Не авторизован"
// @Failure 403 {object} response.Response "Доступ Запрещен"
// @Failure 500 {object} response.Response "Ошибка сервера"
// @Router /orders [get]
// @Security BearerAuth
func New(log *slog.Logger, getter OrderGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "httpserver.ordersStruct.getOrdersForUser.New"
		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())))

		userID := r.Context().Value("userID").(int32)

		ordersInfo, err := getter.GetFullOrdersByUserID(r.Context(), int32(userID))
		if err != nil {
			response.Error(log, w, r, "No orders", "failed to get the ordersStruct", http.StatusNotFound)
			return
		}

		orders := ordersStruct.MakeOrders(ordersInfo)
		if len(orders) == 0 {
			log.Info("No orders found")
			render.Status(r, http.StatusNoContent)
			render.JSON(w, r, Response{
				[]ordersStruct.Order{},
			})
		}

		log.Info("got orders")
		render.Status(r, http.StatusOK)
		render.JSON(w, r, Response{
			orders,
		})
	}
}
