package order

import (
	"context"

	"frappuccino/internal/entity"
)

type orderRepo interface {
	CreateOrder(ctx context.Context, order entity.Order, items []entity.OrderItem) (string, error)
	GetMenuItemPrice(ctx context.Context, menuItemID string) (float64, error)
	GetOrderByID(ctx context.Context, orderID string) (entity.Order, error)
	GetOrderItemsByOrderID(ctx context.Context, orderID string) ([]entity.OrderItem, error)
	GetAllOrders(ctx context.Context) ([]entity.Order, error)
	UpdateOrder(ctx context.Context, orderID string, updates map[string]interface{}) error
	GetAllOrderStatusHistory(ctx context.Context) ([]entity.OrderStatusHistory, error)
	DeleteOrder(ctx context.Context, id string) (string, error)
}

// type ServiceInterface interface {
// 	CreateOrder(ctx context.Context, req orderdto.CreateOrderRequest) (string, error)
// 	GetOrderByID(ctx context.Context, id string) (orderdto.GetOrderResponse, error)
// }
