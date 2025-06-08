package getOrderByID

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/yourgfslove/GodFoodApi/internal/database"
	"github.com/yourgfslove/GodFoodApi/internal/lib/api/response"
	"github.com/yourgfslove/GodFoodApi/internal/lib/logger/sl"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

type orderGetter interface {
	GetFullOrderByID(ctx context.Context, id int32) ([]database.GetFullOrderByIDRow, error)
}

type Response struct {
	response.Response
	RestaurantName    string  `json:"restaurant_Name"`
	RestaurantAddress string  `json:"restaurant_Address"`
	RestaurantPhone   string  `json:"restaurant_Phone"`
	DeliveryAddress   string  `json:"delivery_Address"`
	CourierName       string  `json:"courierName"`
	UserName          string  `json:"user_name"`
	Status            string  `json:"status"`
	CreatedAt         string  `json:"created_at"`
	Items             []item  `json:"items"`
	TotalPrice        float64 `json:"total_price"`
}
type item struct {
	ItemName  string  `json:"item_name"`
	ItemPrice float64 `json:"price"`
	Quantity  int32   `json:"quantity"`
}

func New(log *slog.Logger, getter orderGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http_server.orders.getOrderByID.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())))

		userID := r.Context().Value("userID").(int32)

		orderID := chi.URLParam(r, "id")
		parsedOrderID, err := strconv.ParseInt(orderID, 10, 32)
		if err != nil {
			log.Info("invalid user id", sl.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid user id"))
			return
		}

		order, err := getter.GetFullOrderByID(r.Context(), int32(parsedOrderID))
		if err != nil {
			log.Info("invalid user id", sl.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid user id"))
			return
		}

		if len(order) == 0 {
			log.Info("no order for following id")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("no order for following id"))
			return
		}

		if order[0].CustomerID != userID {
			log.Info("access denied")
			w.WriteHeader(http.StatusForbidden)
			render.JSON(w, r, response.Error("access denied"))
			return
		}

		resp := Response{
			Response:          response.OK(),
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
		log.Info("got order")
		render.JSON(w, r, resp)
	}
}
