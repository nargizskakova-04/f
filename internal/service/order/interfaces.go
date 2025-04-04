package order

import (
	"context"
	"time"

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
	GetNumberOfOrderedItems(ctx context.Context, startDate, endDate *time.Time) (map[string]int, error)
}

//	type ServiceInterface interface {
//		CreateOrder(ctx context.Context, req orderdto.CreateOrderRequest) (string, error)
//		GetOrderByID(ctx context.Context, id string) (orderdto.GetOrderResponse, error)
//	}
//
// menuRepo defines methods for working with menu items and ingredients
type menuRepo interface {
	GetMenuItemIngredients(ctx context.Context, menuItemID string) ([]entity.MenuItemIngredient, error)
	GetMenuItem(ctx context.Context) ([]entity.MenuItem, error)
}

// inventoryRepo defines methods for working with inventory
type inventoryRepo interface {
	GetInventoryByID(ctx context.Context, id string) (entity.Inventory, error)
	UpdateInventory(ctx context.Context, updates map[string]interface{}, id string) (string, error)
	CreateInventoryTransaction(ctx context.Context, transaction entity.InventoryTransaction) error
}
