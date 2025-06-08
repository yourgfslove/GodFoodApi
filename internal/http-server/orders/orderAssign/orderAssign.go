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
)

type StatusUpdater interface {
}

type StatusGetter interface {
	GetOrderStatusByID(ctx context.Context, id int32) (string, error)
}

type courierGetter interface {
	GetUserByID(ctx context.Context, id int32) (database.User, error)
}

type currentOrderGetter interface {
	GetCurrentOrderForCourier(ctx context.Context, courierid sql.NullInt32) (int32, error)
}

func New(
	log *slog.Logger,
	getterStatus StatusGetter,
	updater StatusUpdater,
	getterCourier courierGetter,
	getterCurrent currentOrderGetter,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http_server.orders.ordersAssign.New"
		log = log.With(slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())))

		userID := r.Context().Value("usersID").(int32)

		courierInfo, err := getterCourier.GetUserByID(r.Context(), userID)
		if err != nil {
			log.Error("cannot get courier")
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("Cannot get courier"))
			return
		}

		if courierInfo.UserRole != "courier" {
			w.WriteHeader(http.StatusForbidden)
			render.JSON(w, r, response.Error("Access denied"))
			return
		}

		orderID := chi.URLParam(r, "id")

		if orderID == "" {
			log.Info("no order id provided")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("Missing orderID"))
			return
		}

		parsedOrderID, err := strconv.ParseInt(orderID, 10, 32)
		if err != nil {
			log.Info("cannot parse orderID")
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("Cannot parse orderID"))
			return
		}

		orderInfo, err := getterStatus.GetOrderStatusByID(r.Context(), int32(parsedOrderID))
		if err != nil {
			log.Error("cannot get order")
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("Cannot get order"))
			return
		}

		if len(orderInfo) == 0 {
			w.WriteHeader(http.StatusNotFound)
			render.JSON(w, r, response.Error("order not found"))
			return
		}

		if orderInfo != "pending" {
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("Order already accepted"))
			return
		}

		currentOrder, err := getterCurrent.GetCurrentOrderForCourier(r.Context(), sql.NullInt32{
			Int32: int32(courierInfo.ID),
		})
		if err != nil {
			log.Error("cannot get current order")
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("Cannot get current order"))
			return
		}

	}
}
