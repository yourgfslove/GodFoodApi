package placeorder

import (
	"context"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/yourgfslove/GodFoodApi/internal/database"
	"github.com/yourgfslove/GodFoodApi/internal/lib/api/response"
	"github.com/yourgfslove/GodFoodApi/internal/lib/auth/JWT"
	"github.com/yourgfslove/GodFoodApi/internal/lib/auth/getToken"
	"github.com/yourgfslove/GodFoodApi/internal/lib/logger/sl"
	"log/slog"
	"net/http"
)

type orderCreater interface {
	CreateOrder(ctx context.Context, arg database.CreateOrderParams) (database.Order, error)
}

type menuItemsAdder interface {
	AddItems(ctx context.Context, arg database.AddItemsParams) ([]database.Orderitem, error)
}

type userGetter interface {
	GetUserByID(ctx context.Context, id int32) (database.User, error)
}

type Request struct {
	RestaurantID int32  `json:"restaurant_id"`
	Address      string `json:"address,omitempty"`
	Items        []struct {
		MenuitemID int32 `json:"menuitem_id"`
		Quantity   int32 `json:"quantity"`
	} `json:"items"`
}

type Response struct {
	response.Response
	OrderID      int32  `json:"order_id"`
	RestaurantID int32  `json:"restaurant_id"`
	Status       string `json:"status"`
	CreatedAt    string `json:"created_at"`
	Address      string `json:"user_address"`
	Items        []struct {
		MenuitemID int32 `json:"menuitem_id"`
		Quantity   int32 `json:"quantity"`
	} `json:"items"`
}

func New(log *slog.Logger, creater orderCreater, userGetter userGetter, adder menuItemsAdder, tokensecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.ordersStruct.placeorder"
		log = log.With(slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())))
		token, err := getToken.GetTokenFromHeader(r.Header, "Bearer")
		if err != nil {
			log.Info("failed to get token from header", "err", sl.Err(err))
			w.WriteHeader(http.StatusUnauthorized)
			render.JSON(w, r, response.Error("failed to get token"))
			return
		}
		userID, err := JWT.ValidateJWT(token, tokensecret)
		if err != nil {
			log.Info("failed to validate token", "err", sl.Err(err))
			w.WriteHeader(http.StatusUnauthorized)
			render.JSON(w, r, response.Error("failed to validate token"))
			return
		}
		userInfo, err := userGetter.GetUserByID(r.Context(), int32(userID))
		if err != nil {
			log.Info("failed to get user by id", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to get user"))
			return
		}
		var req Request
		if err := render.DecodeJSON(r.Body, &req); err != nil {
			log.Info("failed to decode request body", sl.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("failed to decode request body"))
			return
		}
		if len(req.Items) == 0 {
			log.Info("no items found")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("no items found"))
			return
		}
		address := req.Address
		if address == "" {
			if !userInfo.Address.Valid {
				log.Info("No user address")
				w.WriteHeader(http.StatusBadRequest)
				render.JSON(w, r, response.Error("no user address"))
				return
			}
			address = userInfo.Address.String
		}
		order, err := creater.CreateOrder(r.Context(), database.CreateOrderParams{
			Customerid:   userInfo.ID,
			Restaurantid: req.RestaurantID,
			Address:      address,
		})
		if err != nil {
			log.Info("failed to create order", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to create order"))
			return
		}
		orderIDs := make([]int32, len(req.Items))
		itemIDs := make([]int32, len(req.Items))
		quantity := make([]int32, len(req.Items))
		for i, item := range req.Items {
			orderIDs[i] = order.ID
			itemIDs[i] = item.MenuitemID
			quantity[i] = item.Quantity
		}
		_, err = adder.AddItems(r.Context(), database.AddItemsParams{
			Column1: orderIDs,
			Column2: itemIDs,
			Column3: quantity,
		})
		if err != nil {
			log.Info("failed to add items", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to add items"))
			return
		}
		log.Info("successfully added items")
		w.WriteHeader(http.StatusCreated)
		render.JSON(w, r, Response{
			response.OK(),
			order.ID,
			req.RestaurantID,
			order.Status,
			order.CreatedAt.Time.Format("2006-01-02 15:04:05"),
			order.Address,
			req.Items,
		})
	}
}
