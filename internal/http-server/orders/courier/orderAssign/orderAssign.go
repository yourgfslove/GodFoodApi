package orderAssign

import (
	"context"
	"database/sql"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/yourgfslove/GodFoodApi/internal/database"
	"github.com/yourgfslove/GodFoodApi/internal/lib/api/response"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

type Response struct {
	RestaurantName    string  `json:"restaurant_Name" example:"Mac"`
	RestaurantAddress string  `json:"restaurant_Address" example:"123 address"`
	RestaurantPhone   string  `json:"restaurant_Phone" example:"89056666666"`
	DeliveryAddress   string  `json:"delivery_Address" example:"1222 address"`
	CourierName       string  `json:"courierName" example:"Bill"`
	UserName          string  `json:"user_name" example:"Ivan"`
	Status            string  `json:"status" example:"pending"`
	CreatedAt         string  `json:"created_at" example:"2020-01-01T00:00:00+09:00"`
	Items             []item  `json:"items"`
	TotalPrice        float64 `json:"total_price" example:"300.0"`
}
type item struct {
	ItemName  string  `json:"item_name" example:"Burger with cheese"`
	ItemPrice float64 `json:"price" example:"100.0"`
	Quantity  int32   `json:"quantity" example:"3"`
}

const StatusPending = "pending"

type StatusUpdater interface {
	UpdateCourierID(ctx context.Context, arg database.UpdateCourierIDParams) ([]database.UpdateCourierIDRow, error)
}

type StatusGetter interface {
	GetOrderStatusByID(ctx context.Context, id int32) (string, error)
}

type courierGetter interface {
	GetUserByID(ctx context.Context, id int32) (database.User, error)
}

type currentOrderGetter interface {
	GetCurrentIDOrderForCourier(ctx context.Context, courierid sql.NullInt32) (int32, error)
}

// Orders godoc
// @Summary Взятие заказа курьером
// @Description Назначает заказ на курьера
// @Tags Orders
// @Accept json
// @Produce json
// @Param id path int true "ID Заказа"
// @Success 200 {object} orderAssign.Response "Заказ назначен"
// @Failure 400 {object} response.Response "Некорректные данные"
// @Failure 401 {object} response.Response "Неавторизован"
// @Failure 404 {object} response.Response "Заказ не найден"
// @Failure 500 {object} response.Response "Серверная Ошибка"
// @Router /orders/{id}/assign [patch]
// @Security BearerAuth
func New(
	log *slog.Logger,
	getterStatus StatusGetter,
	updater StatusUpdater,
	getterCourier courierGetter,
	getterCurrent currentOrderGetter,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http_server.orders.courier.ordersAssign.New"
		log = log.With(slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())))

		userID := r.Context().Value("userID").(int32)

		courierInfo, err := getterCourier.GetUserByID(r.Context(), userID)
		if err != nil {
			response.Error(log, w, r, "Can not get courier", "Can not get courier", http.StatusInternalServerError)
			return
		}

		if courierInfo.UserRole != "courier" {
			response.Error(log, w, r, "Access denied", "wrong role", http.StatusForbidden)
			return
		}

		orderID := chi.URLParam(r, "id")

		if orderID == "" {
			response.Error(log, w, r, "Missing orderID", "no order id provided", http.StatusBadRequest)
			return
		}

		parsedOrderID, err := strconv.ParseInt(orderID, 10, 32)
		if err != nil {
			response.Error(log, w, r, "Cannot parse orderID", "Cannot parse orderID", http.StatusInternalServerError)
			return
		}

		if parsedOrderID < 0 {
			response.Error(log, w, r, "Wrong orderID", "Invalid orderID", http.StatusBadRequest)
			return
		}

		orderInfo, err := getterStatus.GetOrderStatusByID(r.Context(), int32(parsedOrderID))
		if err != nil {
			response.Error(log, w, r, "No order", "cannot get order", http.StatusNotFound)
			return
		}

		if orderInfo != StatusPending {
			response.Error(log, w, r, "Order already accepted", "order already accepted", http.StatusForbidden)
			return
		}

		_, err = getterCurrent.GetCurrentIDOrderForCourier(r.Context(), sql.NullInt32{
			Int32: int32(courierInfo.ID),
			Valid: true,
		})

		if err == nil {
			response.Error(log, w, r, "already have order", "already have order", http.StatusForbidden)
			return
		}

		order, err := updater.UpdateCourierID(r.Context(), database.UpdateCourierIDParams{
			Courierid: sql.NullInt32{Int32: courierInfo.ID, Valid: true},
			ID:        int32(parsedOrderID),
		})
		if err != nil {
			response.Error(log, w, r, "Can not update order", "Can not update order", http.StatusInternalServerError)
			return
		}

		if len(order) == 0 {
			response.Error(log, w, r, "order not found", "no order updated", http.StatusNotFound)
			return
		}

		resp := Response{
			RestaurantName:    order[0].RestaurantName.String,
			RestaurantAddress: order[0].RestaurantAddress.String,
			RestaurantPhone:   order[0].RestaurantPhone,
			DeliveryAddress:   order[0].DeliveryAddress,
			CourierName:       order[0].CourierName.String,
			UserName:          order[0].CostomerName.String,
			Status:            order[0].Status,
			CreatedAt:         order[0].CreatedAt.Time.Format(time.RFC1123),
			Items:             []item{},
			TotalPrice:        0,
		}
		for _, v := range order {
			resp.Items = append(resp.Items, item{
				ItemName:  v.MenuItemName,
				ItemPrice: v.Price,
				Quantity:  v.Quanity,
			})
			resp.TotalPrice += v.Price * float64(v.Quanity)
		}

		log.Info("order assigned")
		render.Status(r, http.StatusOK)
		render.JSON(w, r, resp)
	}
}
