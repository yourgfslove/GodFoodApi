package getOrderByID

import (
	"context"
	"database/sql"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/yourgfslove/GodFoodApi/internal/database"
	"github.com/yourgfslove/GodFoodApi/internal/lib/api/response"
	"github.com/yourgfslove/GodFoodApi/internal/lib/auth/JWT"
	"github.com/yourgfslove/GodFoodApi/internal/lib/auth/getToken"
	"github.com/yourgfslove/GodFoodApi/internal/lib/logger/sl"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

type orderGetter interface {
	GetFullOrderByID(ctx context.Context, id int32) ([]database.GetFullOrderByIDRow, error)
}

type courierNameGetter interface {
	GetNameByID(ctx context.Context, id int32) (sql.NullString, error)
}

type Response struct {
	response.Response
	OrderID      int32   `json:"order_id"`
	RestaurantID int32   `json:"restaurant_id"`
	CourierName  string  `json:"courier_name"`
	Status       string  `json:"status"`
	UserAddress  string  `json:"user_address"`
	CreatedAt    string  `json:"created_at"`
	Items        []item  `json:"items"`
	TotalPrice   float64 `json:"total_price"`
}
type item struct {
	ItemName  string  `json:"item_name"`
	ItemPrice float64 `json:"price"`
	Quantity  int32   `json:"quantity"`
}

func New(log *slog.Logger, getter orderGetter, courierNameGetter courierNameGetter, tokenSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http_server.orders.getOrderByID.New"
		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())))
		token, err := getToken.GetTokenFromHeader(r.Header, "Bearer")
		if err != nil {
			log.Info("failed to get token", sl.Err(err))
			w.WriteHeader(http.StatusUnauthorized)
			render.JSON(w, r, response.Error("failed to get token"))
			return
		}
		userID, err := JWT.ValidateJWT(token, tokenSecret)
		if err != nil {
			log.Info("failed to validate token", sl.Err(err))
			w.WriteHeader(http.StatusUnauthorized)
			render.JSON(w, r, response.Error("failed to validate token"))
			return
		}
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
		if order[0].Customerid != userID {
			log.Info("access denied")
			w.WriteHeader(http.StatusForbidden)
			render.JSON(w, r, response.Error("access denied"))
			return
		}

		courierName, err := courierNameGetter.GetNameByID(r.Context(), order[0].Courierid.Int32)
		if err != nil {
			log.Info("failed to get courier name", sl.Err(err))
		}

		resp := Response{
			Response:     response.OK(),
			OrderID:      order[0].OrderID,
			RestaurantID: order[0].OrderRestaurantID,
			CourierName:  courierName.String,
			Status:       order[0].Status,
			UserAddress:  order[0].Address,
			CreatedAt:    order[0].CreatedAt.Time.Format(time.RFC3339),
			Items:        make([]item, 0, len(order)),
			TotalPrice:   0,
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
