package entity

import (
	"encoding/json"
	"time"
)

type Order struct {
	OrderID             string          `json:"order_id"`
	CustomerName        string          `json:"customer_name"`
	SpecialInstructions json.RawMessage `json:"special_instructions,omitempty"` // JSONB
	TotalAmount         float64         `json:"total_amount"`
	Status              string          `json:"status"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
}

type OrderItem struct {
	OrderItemID    string          `json:"order_item_id"`
	OrderID        string          `json:"order_id"`
	MenuItemID     string          `json:"menu_item_id"`
	Quantity       int             `json:"quantity"`
	PriceAtTime    float64         `json:"price_at_time"`
	Customizations json.RawMessage `json:"customizations,omitempty"` // JSONB
}

// Add this to entity package
type OrderStatusHistory struct {
	OrderStatusID string    `json:"order_status_id"`
	OrderID       string    `json:"order_id"`
	OldStatus     string    `json:"old_status"`
	NewStatus     string    `json:"new_status"`
	ChangedAt     time.Time `json:"changed_at"`
	ChangeReason  string    `json:"change_reason"`
}
