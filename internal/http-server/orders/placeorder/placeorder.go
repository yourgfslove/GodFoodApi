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
	RestaurantID int32  `json:"restaurant_id" example:"14"`
	Address      string `json:"address,omitempty" example:"123 address"`
	Items        []struct {
		MenuitemID int32 `json:"menuitem_id" example:"6"`
		Quantity   int32 `json:"quantity" example:"5"`
	} `json:"items"`
}

type Response struct {
	OrderID      int32  `json:"order_id" example:"12"`
	RestaurantID int32  `json:"restaurant_id" example:"14"`
	Status       string `json:"status" example:"pending"`
	CreatedAt    string `json:"created_at" example:"Tue, 17 Jun 2025 00:25:16 +0000"`
	Address      string `json:"user_address" example:"123 address"`
	Items        []struct {
		MenuitemID int32 `json:"menuitem_id" example:"6"`
		Quantity   int32 `json:"quantity" example:"5"`
	} `json:"items"`
}

// Orders godoc
// @Summary Создание нового заказа авторизованным пользователем
// @Description Создает новый заказ
// @Tags Orders
// @Accept json
// @Produce json
// @Param request body placeorder.Request true "Данные для добавления"
// @Success 200 {object} placeorder.Response "Новый заказ успешно создан "
// @Failure 400 {object} response.Response "Некорректные данные"
// @Failure 401 {object} response.Response "Неавторизован"
// @Failure 500 {object} response.Response "Серверная Ошибка"
// @Router /orders [post]
// @Security BearerAuth
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
			response.Error(log, w, r, "failed to get user", sl.Err(err).String(), http.StatusInternalServerError)
			return
		}

		var req Request
		if err := render.DecodeJSON(r.Body, &req); err != nil {
			response.Error(log, w, r, "failed to decode request body", sl.Err(err).String(), http.StatusBadRequest)
			return
		}

		if len(req.Items) == 0 {
			response.Error(log, w, r, "Not Found", "not items found", http.StatusNotFound)
			return
		}

		availableItems, err := availableGetter.GetAvailableIDByRestaurantID(r.Context(), req.RestaurantID)
		if err != nil {
			response.Error(log, w, r, "failed to get available items", "failed to find available items", http.StatusInternalServerError)
			return
		}
		for _, item := range req.Items {
			if !slices.Contains(availableItems, item.MenuitemID) {
				response.Error(log, w, r,
					fmt.Sprintf("item %v is not available", item.MenuitemID),
					fmt.Sprintf("item %v is not available", item.MenuitemID),
					http.StatusBadRequest)
				return
			}
		}
		address := req.Address
		if address == "" {
			if !userInfo.Address.Valid {
				response.Error(log, w, r, "no user address", "No address", http.StatusBadRequest)
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
			response.Error(log, w, r, "something went wrong", "failed to create order", http.StatusInternalServerError)
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
			response.Error(log, w, r, "something went wrong", "failed to add items", http.StatusInternalServerError)
			return
		}
		log.Info("successfully added items")

		render.Status(r, http.StatusCreated)
		render.JSON(w, r, Response{
			order.ID,
			req.RestaurantID,
			order.Status,
			order.CreatedAt.Time.Format(time.RFC3339),
			order.Address,
			req.Items,
		})
	}
}
