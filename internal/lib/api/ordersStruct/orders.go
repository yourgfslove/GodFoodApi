package ordersStruct

import (
	"github.com/yourgfslove/GodFoodApi/internal/database"
	"time"
)

type Order struct {
	RestaurantName    string  `json:"restaurant_Name"`
	RestaurantAddress string  `json:"restaurant_Address"`
	RestaurantPhone   string  `json:"restaurant_Phone"`
	DeliveryAddress   string  `json:"delivery_Address"`
	UserName          string  `json:"user_name"`
	Status            string  `json:"status"`
	TotalPrice        float64 `json:"total_price"`
	CreatedAt         string  `json:"created_at"`
	Items             []Item  `json:"items"`
}

type OrderForCourier struct {
	RestaurantName    string  `json:"restaurant_Name"`
	RestaurantAddress string  `json:"restaurant_Address"`
	RestaurantPhone   string  `json:"restaurant_Phone"`
	DeliveryAddress   string  `json:"delivery_Address"`
	UserPhone         string  `json:"user_phone"`
	Reward            float64 `json:"reward"`
	CreatedAt         string  `json:"created_at"`
	Items             []Item  `json:"items"`
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
				RestaurantName:    row.RestaurantName.String,
				RestaurantAddress: row.RestaurantAddress.String,
				RestaurantPhone:   row.RestaurantPhone,
				DeliveryAddress:   row.DeliveryAddress,
				UserName:          row.CostomerName.String,
				Status:            row.Status,
				TotalPrice:        0,
				CreatedAt:         row.CreatedAt.Time.Format(time.RFC1123),
				Items:             []Item{},
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

func MakePendingOrders(rows []database.GetFullPendingOrdersRow) []OrderForCourier {
	ordersMap := make(map[int32]*OrderForCourier, len(rows))
	for _, row := range rows {
		order, exists := ordersMap[row.OrderID]
		if !exists {
			order = &OrderForCourier{
				RestaurantName:    row.RestaurantName.String,
				RestaurantAddress: row.RestaurantAddress.String,
				RestaurantPhone:   row.RestaurantPhone,
				DeliveryAddress:   row.DeliveryAddress,
				UserPhone:         row.CustomerPhone,
				Reward:            0,
				CreatedAt:         row.CreatedAt.Time.Format(time.RFC1123),
				Items:             []Item{},
			}
			ordersMap[row.OrderID] = order
		}
		order.Items = append(order.Items, Item{
			MenuItemID: row.MenuItemID,
			ItemName:   row.MenuItemName,
			ItemPrice:  row.Price,
			Quantity:   row.Quanity,
		})
		order.Reward += row.Price * float64(row.Quanity) * 0.05
	}
	var orders []OrderForCourier
	for _, order := range ordersMap {
		orders = append(orders, *order)
	}
	return orders
}
