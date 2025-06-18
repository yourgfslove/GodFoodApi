package getPendingOrders

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

type ordersGetter interface {
	GetFullPendingOrders(ctx context.Context) ([]database.GetFullPendingOrdersRow, error)
}

type userGetter interface {
	GetUserByID(ctx context.Context, id int32) (database.User, error)
}

type Response struct {
	PendingOrders []ordersStruct.OrderForCourier `json:"pending_orders"`
}

// Orders godoc
// @Summary Получение всех доступных для доставки заказов
// @Description Возвращает все заказы, которые еще не взяты
// @Tags Orders
// @Accept json
// @Produce json
// @Success 200 {object} getPendingOrders.Response "Заказы успешно получен"
// @Failure 400 {object} response.Response "Некорректные данные"
// @Failure 401 {object} response.Response "Неавторизован"
// @Failure 404 {object} response.Response "Заказы не найдены"
// @Failure 500 {object} response.Response "Ошибка сервера"
// @Router /orders/pending [get]
// @Security BearerAuth
func New(log *slog.Logger, ordersGetter ordersGetter, userGetter userGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.orders.courier.getPendingOrders"
		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())))

		userID := r.Context().Value("userID").(int32)

		userInfo, err := userGetter.GetUserByID(r.Context(), userID)
		if err != nil {
			response.Error(log, w, r, "failed to get user", "No user", http.StatusInternalServerError)
			return
		}

		if userInfo.UserRole != "courier" {
			response.Error(log, w, r, "access denied", "Wrong role", http.StatusForbidden)
			return
		}

		orders, err := ordersGetter.GetFullPendingOrders(r.Context())
		if err != nil {
			response.Error(log, w, r, "failed to get pending orders", "no pending orders", http.StatusInternalServerError)
			return
		}

		respOrders := ordersStruct.MakePendingOrders(orders)
		render.Status(r, http.StatusOK)
		render.JSON(w, r, Response{
			respOrders,
		})
	}
}
