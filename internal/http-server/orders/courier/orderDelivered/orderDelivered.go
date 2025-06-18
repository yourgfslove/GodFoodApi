package orderDelivered

import (
	"context"
	"database/sql"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/yourgfslove/GodFoodApi/internal/database"
	"github.com/yourgfslove/GodFoodApi/internal/lib/api/response"
	"log/slog"
	"net/http"
)

type courierGetter interface {
	GetUserByID(ctx context.Context, id int32) (database.User, error)
}

type statusUpdater interface {
	UpdateOrderStatus(ctx context.Context, arg database.UpdateOrderStatusParams) error
}

type currentOrderGetter interface {
	GetCurrentOrderForCourier(ctx context.Context, courierid sql.NullInt32) ([]database.GetCurrentOrderForCourierRow, error)
}

type Response struct {
	OrderID int32 `json:"order_id"`
}

// Orders godoc
// @Summary Изменение статуса заказа
// @Description Изменяет статус доставляемого заказа
// @Tags Orders
// @Accept json
// @Produce json
// @Param id path int true "ID Заказа"
// @Success 200 {object} orderDelivered.Response "Заказ доставлен"
// @Failure 400 {object} response.Response "Некорректные данные"
// @Failure 401 {object} response.Response "Неавторизован"
// @Failure 403 {object} response.Response "Доступ запрещен"
// @Failure 404 {object} response.Response "Заказ не найден"
// @Failure 500 {object} response.Response "Серверная Ошибка"
// @Router /orders/delivered [patch]
// @Security BearerAuth
func New(log *slog.Logger, getterCourier courierGetter, updater statusUpdater, getterOrder currentOrderGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http_server.orders.courier.orderDelivered.New"
		log = log.With(slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())))

		userID := r.Context().Value("userID").(int32)

		order, err := getterOrder.GetCurrentOrderForCourier(r.Context(), sql.NullInt32{Int32: userID, Valid: true})
		if err != nil {
			response.Error(log, w, r, "Not found", "current order not found", http.StatusNotFound)
			return
		}

		if len(order) == 0 {
			response.Error(log, w, r, "Not found", "Empty order", http.StatusNotFound)
			return
		}

		if err := updater.UpdateOrderStatus(r.Context(), database.UpdateOrderStatusParams{
			Courierid: sql.NullInt32{Int32: userID, Valid: true},
			ID:        order[0].OrderID,
		}); err != nil {
			response.Error(log, w, r, "Failed to update order", "cannot update", http.StatusInternalServerError)
			return
		}

		render.Status(r, http.StatusOK)
		render.JSON(w, r, Response{
			order[0].OrderID,
		})
	}
}
