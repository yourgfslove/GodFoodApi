package placeorder

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/yourgfslove/GodFoodApi/internal/database"
	"github.com/yourgfslove/GodFoodApi/internal/lib/api/response"
	"github.com/yourgfslove/GodFoodApi/internal/lib/logger/sl"
	"log/slog"
	"net/http"
	"slices"
	"time"
)

type orderCreater interface {
	CreateOrder(ctx context.Context, arg database.CreateOrderParams) (database.Order, error)
}

type menuItemsAdderNGetter interface {
	AddItems(ctx context.Context, arg database.AddItemsParams) ([]database.Orderitem, error)
}

type availableItemsGetter interface {
	GetAvailableIDByRestaurantID(ctx context.Context, restaurantID int32) ([]int32, error)
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

func New(
	log *slog.Logger,
	creater orderCreater,
	userGetter userGetter,
	adder menuItemsAdderNGetter,
	availableGetter availableItemsGetter,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.ordersStruct.placeorder"
		log = log.With(slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())))

		userID := r.Context().Value("userID").(int32)

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

		availableItems, err := availableGetter.GetAvailableIDByRestaurantID(r.Context(), req.RestaurantID)
		if err != nil {
			log.Info("failed to get available items", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to get available items"))
			return
		}
		for _, item := range req.Items {
			if !slices.Contains(availableItems, item.MenuitemID) {
				log.Info(fmt.Sprintf("item %v is not available", item.MenuitemID))
				w.WriteHeader(http.StatusBadRequest)
				render.JSON(w, r, response.Error(fmt.Sprintf("item %v is not available", item.MenuitemID)))
				return
			}
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
			order.CreatedAt.Time.Format(time.RFC850),
			order.Address,
			req.Items,
		})
	}
}
