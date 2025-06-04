package ordersStruct

import (
	"github.com/yourgfslove/GodFoodApi/internal/database"
	"strconv"
	"time"
)

type Order struct {
	RestaurantID  string  `json:"restaurant_id"`
	Status        string  `json:"status"`
	TotalPrice    float64 `json:"total_price"`
	ClientAddress string  `json:"client_address"`
	CreatedAt     string  `json:"created_at"`
	Items         []Item  `json:"items"`
}

type Item struct {
	MenuItemID int32   `json:"menu_item_id"`
	ItemName   string  `json:"item_name"`
	ItemPrice  float64 `json:"item_price"`
	Quantity   int32   `json:"quantity"`
}

func MakeOrders(rows []database.GetFullOrdersByUserIDRow) []Order {
	ordersMap := make(map[int32]*Order, len(rows))
	for _, row := range rows {
		order, exists := ordersMap[row.OrderID]
		if !exists {
			order = &Order{
				RestaurantID:  strconv.Itoa(int(row.OrderRestaurantID)),
				Status:        row.Status,
				TotalPrice:    0,
				ClientAddress: row.Address,
				CreatedAt:     row.CreatedAt.Time.Format(time.RFC822),
				Items:         []Item{},
			}
			ordersMap[row.OrderID] = order
		}
		order.Items = append(order.Items, Item{
			MenuItemID: row.MenuItemID,
			ItemName:   row.MenuItemName,
			ItemPrice:  row.Price,
			Quantity:   row.Quanity,
		})
		order.TotalPrice += row.Price * float64(row.Quanity)
	}
	var orders []Order
	for _, order := range ordersMap {
		orders = append(orders, *order)
	}
	return orders
}
