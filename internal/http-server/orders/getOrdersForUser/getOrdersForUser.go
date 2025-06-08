package getOrdersForUser

import (
	"context"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/yourgfslove/GodFoodApi/internal/database"
	"github.com/yourgfslove/GodFoodApi/internal/lib/api/ordersStruct"
	"github.com/yourgfslove/GodFoodApi/internal/lib/api/response"
	"github.com/yourgfslove/GodFoodApi/internal/lib/logger/sl"
	"log/slog"
	"net/http"
)

type OrderGetter interface {
	GetFullOrdersByUserID(ctx context.Context, customerid int32) ([]database.GetFullOrdersByUserIDRow, error)
}

type Response struct {
	response.Response
	Orders []ordersStruct.Order `json:"ordersStruct"`
}

func New(log *slog.Logger, getter OrderGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "httpserver.ordersStruct.getOrdersForUser.New"
		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())))

		userID := r.Context().Value("user_id").(int32)

		ordersInfo, err := getter.GetFullOrdersByUserID(r.Context(), int32(userID))
		if err != nil {
			log.Info("failed to get the ordersStruct", sl.Err(err))
			w.WriteHeader(http.StatusUnauthorized)
			render.JSON(w, r, response.Error("failed to get the ordersStruct"))
			return
		}
		orders := ordersStruct.MakeOrders(ordersInfo)
		if len(orders) == 0 {
			log.Info("No orders found")
			w.WriteHeader(http.StatusNoContent)
			render.JSON(w, r, Response{
				response.OK(),
				[]ordersStruct.Order{},
			})
		}
		log.Info("got orders")
		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, Response{
			response.OK(),
			orders,
		})
	}
}
