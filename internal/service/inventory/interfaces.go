package inventory

import (
	"context"

	"frappuccino/internal/entity"
)

type inventoryRepo interface {
	CreateInventory(ctx context.Context, inventory entity.Inventory) (string, error)
	GetInventory(ctx context.Context) ([]entity.Inventory, error)
	GetInventoryByID(ctx context.Context, id string) (entity.Inventory, error)
	DeleteInventory(ctx context.Context, id string) (string, error)
	UpdateInventory(ctx context.Context, updates map[string]interface{}, id string) (string, error)
	CreateInventoryTransaction(ctx context.Context, transaction entity.InventoryTransaction) error
	GetInventoryTransactions(ctx context.Context, ingredientID string) ([]entity.InventoryTransaction, error)
	GetLeftOvers(ctx context.Context, sortBy string, page, pageSize int) ([]entity.Inventory, int, error)
}
