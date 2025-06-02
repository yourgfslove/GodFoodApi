package getOrdersForUser

import (
	"context"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/yourgfslove/GodFoodApi/internal/database"
	"github.com/yourgfslove/GodFoodApi/internal/lib/api/ordersStruct"
	"github.com/yourgfslove/GodFoodApi/internal/lib/api/response"
	"github.com/yourgfslove/GodFoodApi/internal/lib/auth/JWT"
	"github.com/yourgfslove/GodFoodApi/internal/lib/auth/getToken"
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

func New(log *slog.Logger, getter OrderGetter, tokenSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "httpserver.ordersStruct.getOrdersForUser.New"
		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())))
		token, err := getToken.GetTokenFromHeader(r.Header, "Bearer")
		if err != nil {
			log.Info("failed to get the token", sl.Err(err))
			w.WriteHeader(http.StatusUnauthorized)
			render.JSON(w, r, response.Error("failed to get token"))
		}
		userID, err := JWT.ValidateJWT(token, tokenSecret)
		if err != nil {
			log.Info("failed to validate the token", sl.Err(err))
			w.WriteHeader(http.StatusUnauthorized)
			render.JSON(w, r, response.Error("failed to validate the token"))
			return
		}
		ordersInfo, err := getter.GetFullOrdersByUserID(r.Context(), int32(userID))
		if err != nil {
			log.Info("failed to get the ordersStruct", sl.Err(err))
			w.WriteHeader(http.StatusUnauthorized)
			render.JSON(w, r, response.Error("failed to get the ordersStruct"))
			return
		}
		orders := ordersStruct.MakeOrders(ordersInfo)
		if orders == nil {
			log.Info("No orders found")
			w.WriteHeader(http.StatusNoContent)
			render.JSON(w, r, response.Error("No orders found"))
		}
		log.Info("got orders")
		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, Response{
			response.OK(),
			ordersStruct.MakeOrders(ordersInfo),
		})
	}
}
