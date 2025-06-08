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
	response.Response
	PendingOrders []ordersStruct.OrderForCourier `json:"pending_orders"`
}

func New(log *slog.Logger, ordersGetter ordersGetter, userGetter userGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.orders.getPendingOrders"
		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())))

		userID := r.Context().Value("userID").(int32)

		userInfo, err := userGetter.GetUserByID(r.Context(), userID)
		if err != nil {
			log.Info("failed to get the user")
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to get the user"))
			return
		}

		if userInfo.UserRole != "courier" {
			log.Info("access denied")
			w.WriteHeader(http.StatusForbidden)
			render.JSON(w, r, response.Error("access denied"))
			return
		}

		orders, err := ordersGetter.GetFullPendingOrders(r.Context())
		if err != nil {
			log.Info("failed to get the pending orders")
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to get the pending orders"))
			return
		}
		respOrders := ordersStruct.MakePendingOrders(orders)
		render.JSON(w, r, Response{
			response.OK(),
			respOrders,
		})
	}
}
