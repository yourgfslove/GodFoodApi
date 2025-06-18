package ordersStruct

import (
	"github.com/yourgfslove/GodFoodApi/internal/database"
	"time"
)

type Order struct {
	RestaurantName    string  `json:"restaurant_Name" example:"Mac"`
	RestaurantAddress string  `json:"restaurant_Address" example:"123 address"`
	RestaurantPhone   string  `json:"restaurant_Phone" example:"89056666666"`
	DeliveryAddress   string  `json:"delivery_Address" example:"1223 address"`
	UserName          string  `json:"user_name" example:"Ivan"`
	Status            string  `json:"status" example:"pending"`
	TotalPrice        float64 `json:"total_price" example:"300.0"`
	CreatedAt         string  `json:"created_at" example:"2013-08-20T18:08:41+00:00"`
	Items             []Item  `json:"items"`
}

type OrderForCourier struct {
	OrderID           int32   `json:"order_id" example:"1"`
	RestaurantName    string  `json:"restaurant_Name" example:"Mac"`
	RestaurantAddress string  `json:"restaurant_Address" example:"123 address"`
	RestaurantPhone   string  `json:"restaurant_Phone" example:"89056666666"`
	DeliveryAddress   string  `json:"delivery_Address" example:"1223 address"`
	UserPhone         string  `json:"user_phone" example:"Ivan"`
	Reward            float64 `json:"reward" example:"30.0"`
	CreatedAt         string  `json:"created_at" example:"2013-08-20T18:08:41+00:00"`
	Items             []Item  `json:"items"`
}

type Item struct {
	MenuItemID int32   `json:"menu_item_id" example:"1"`
	ItemName   string  `json:"item_name" example:"Burger with cheese"`
	ItemPrice  float64 `json:"item_price" example:"100.0"`
	Quantity   int32   `json:"quantity" example:"3"`
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
				OrderID:           row.OrderID,
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
