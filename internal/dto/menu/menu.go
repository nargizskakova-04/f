package menu

import (
	"encoding/json"
	"time"
)

type MenuItemIngredient struct {
	IngredientID string  `json:"ingredient_id"`
	Quantity     float64 `json:"quantity"`
	Unit         string  `json:"unit"`
}

// CreateMenuItemRequest represents the request for creating a new menu item
// check
type CreateMenuItemRequest struct {
	Name                 string               `json:"name" validate:"required"`
	Description          string               `json:"description"`
	Price                float32              `json:"price" validate:"required,gt=0"`
	Categories           []string             `json:"categories"`
	Allergens            []string             `json:"allergens"`
	Size                 string               `json:"size" validate:"required,oneof=small medium large"`
	CustomizationOptions json.RawMessage      `json:"customization_options"`
	Ingredients          []MenuItemIngredient `json:"ingredients"`
}

type GetMenuResponse struct {
	MenuItemID           string          `json:"menu_item_id"`
	Name                 string          `json:"name"`
	Description          string          `json:"description,omitempty"`
	Price                float32         `json:"price"`
	Categories           []string        `json:"categories"`
	Allergens            []string        `json:"allergens"`
	Size                 string          `json:"size"`
	CustomizationOptions json.RawMessage `json:"customization_options,omitempty"`
	UpdatedAt            time.Time       `json:"updated_at"`
}

type UpdateMenuRequest struct {
	Name                 *string          `json:"name"`
	Description          *string          `json:"description"`
	Price                *float32         `json:"price" validate:"required,gt=0"`
	Categories           *[]string        `json:"categories"`
	Allergens            *[]string        `json:"allergens"`
	Size                 *string          `json:"size" validate:"required,oneof=small medium large"`
	CustomizationOptions *json.RawMessage `json:"customization_options"`
}

type MenuItemIngredientDTO struct {
	IngredientID string  `json:"ingredient_id"`
	Quantity     float64 `json:"quantity"`
	Unit         string  `json:"unit"`
}

type GetPriceHistoryResponse struct {
	ID           string    `json:"id"`
	MenuItemID   string    `json:"menu_item_id"`
	OldPrice     float32   `json:"old_price"`
	NewPrice     float32   `json:"new_price"`
	ChangedAt    time.Time `json:"changed_at"`
	ChangeReason string    `json:"change_reason"`
}
