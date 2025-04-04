package v1

import (
	"context"
	"time"

	"frappuccino/internal/dto/inventory"
	"frappuccino/internal/dto/menu"
	"frappuccino/internal/dto/report"

	orderdto "frappuccino/internal/dto/order"
)

type inventoryInterface interface {
	CreateInventory(ctx context.Context, request inventory.CreateInventoryRequest) (string, error)
	GetInventory(ctx context.Context) ([]inventory.GetInventoryResponse, error)
	GetInventoryByID(ctx context.Context, id string) (inventory.GetInventoryResponse, error)
	DeleteInventory(ctx context.Context, id string) (string, error)
	UpdateInventory(ctx context.Context, request inventory.UpdateInventoryRequest, id string) (string, error)
	RecordInventoryTransaction(ctx context.Context, request inventory.CreateTransactionRequest) error
	GetInventoryTransactions(ctx context.Context, ingredientID string) ([]inventory.TransactionResponse, error)
}

type menuInterface interface {
	CreateMenuItem(ctx context.Context, request menu.CreateMenuItemRequest) (string, error)
	GetMenuItem(ctx context.Context) ([]menu.GetMenuResponse, error)
	GetMenuByID(ctx context.Context, id string) (menu.GetMenuResponse, error)
	DeleteMenu(ctx context.Context, id string) (string, error)
	UpdateMenu(ctx context.Context, request menu.UpdateMenuRequest, id string) (string, error)
	GetAllPriceHistory(ctx context.Context) ([]menu.GetPriceHistoryResponse, error)
}

type orderInterface interface {
	CreateOrder(ctx context.Context, req orderdto.CreateOrderRequest) (string, error)
	GetOrderByID(ctx context.Context, id string) (orderdto.GetOrderResponse, error)
	GetAllOrders(ctx context.Context) ([]orderdto.GetOrderResponse, error)
	UpdateOrder(ctx context.Context, orderID string, req orderdto.UpdateOrderRequest) error
	GetAllOrderStatusHistory(ctx context.Context) ([]orderdto.OrderStatusHistoryResponse, error)
	DeleteOrder(ctx context.Context, id string) (string, error)
	CloseOrder(ctx context.Context, orderID string, reason string) error
	GetNumberOfOrderedItems(ctx context.Context, startDate, endDate *time.Time) (map[string]int, error)
}

type reportInterface interface {
	Search(ctx context.Context, req report.SearchRequest) (report.SearchResponse, error)
	GetOrderedItemsByPeriod(ctx context.Context, req report.OrderedItemsByPeriodRequest) (report.OrderedItemsByPeriodResponse, error)
}
