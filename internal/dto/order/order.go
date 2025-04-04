package order

import (
	"encoding/json"
	"time"
)

type CreateOrderItem struct {
	MenuItemID     string          `json:"menu_item_id"`
	Quantity       int             `json:"quantity"`
	Customizations json.RawMessage `json:"customizations,omitempty"`
}

type CreateOrderRequest struct {
	CustomerName        string            `json:"customer_name"`
	SpecialInstructions json.RawMessage   `json:"special_instructions,omitempty"`
	Items               []CreateOrderItem `json:"items"`
}

type GetOrderItemResponse struct {
	MenuItemID     string          `json:"menu_item_id"`
	Quantity       int             `json:"quantity"`
	PriceAtTime    float64         `json:"price_at_time"`
	Customizations json.RawMessage `json:"customizations,omitempty"`
}

type GetOrderResponse struct {
	OrderID             string                 `json:"order_id"`
	CustomerName        string                 `json:"customer_name"`
	SpecialInstructions json.RawMessage        `json:"special_instructions,omitempty"` // JSONB
	TotalAmount         float64                `json:"total_amount"`
	Status              string                 `json:"status"`
	CreatedAt           time.Time              `json:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at"`
	Items               []GetOrderItemResponse `json:"items"`
}

type UpdateOrderRequest struct {
	CustomerName        *string         `json:"customer_name,omitempty"`
	SpecialInstructions json.RawMessage `json:"special_instructions,omitempty"`
	Status              *string         `json:"status,omitempty"`
	ChangeReason        *string         `json:"change_reason,omitempty"`
}
type OrderStatusHistoryResponse struct {
	OrderStatusID string    `json:"order_status_id"`
	OrderID       string    `json:"order_id"`
	OldStatus     string    `json:"old_status"`
	NewStatus     string    `json:"new_status"`
	ChangedAt     time.Time `json:"changed_at"`
	ChangeReason  string    `json:"change_reason"`
}

// Add to order package
type CloseOrderRequest struct {
	Reason string `json:"reason,omitempty"`
}
