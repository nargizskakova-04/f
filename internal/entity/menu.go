package entity

import (
	"encoding/json"
	"time"
)

type MenuItem struct {
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

type MenuItemIngredient struct {
	ID           string  `json:"id"`
	MenuItemID   string  `json:"menu_item_id"`
	IngredientID string  `json:"ingredient_id"`
	Quantity     float64 `json:"quantity"`
	Unit         string  `json:"unit"`
}

type PriceHistory struct {
	ID           string    `json:"id"`
	MenuItemID   string    `json:"menu_item_id"`
	OldPrice     float32   `json:"old_price"`
	NewPrice     float32   `json:"new_price"`
	ChangedAt    time.Time `json:"changed_at"`
	ChangeReason string    `json:"change_reason"`
}
