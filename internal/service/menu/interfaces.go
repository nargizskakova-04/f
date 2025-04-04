package menu

import (
	"context"
	"frappuccino/internal/entity"
)

type menuRepo interface {
	CreateMenuItem(ctx context.Context, menuItem entity.MenuItem) (string, error)
	GetMenuItem(ctx context.Context) ([]entity.MenuItem, error)
	GetMenuByID(ctx context.Context, id string) (entity.MenuItem, error)
	DeleteMenu(ctx context.Context, id string) (string, error)
	UpdateMenu(ctx context.Context, updates map[string]interface{}, id string) (string, error)
	CreateMenuItemIngredients(ctx context.Context, menuItemID string, ingredients []entity.MenuItemIngredient) error
	GetAllPriceHistory(ctx context.Context) ([]entity.PriceHistory, error)
}
