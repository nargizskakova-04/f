package inventory

import "time"

// DTO = Data Transfer Object

type CreateInventoryRequest struct {
	Name         string  `json:"name"`
	Quantity     float32 `json:"quantity"`
	Unit         string  `json:"unit"`
	UnitPrice    float32 `json:"unit_price"`
	ReorderPoint float32 `json:"reorder_point"`
}

type GetInventoryResponse struct {
	IngredientID string    `json:"ingredient_id"`
	Name         string    `json:"name"`
	Quantity     float32   `json:"quantity"`
	Unit         string    `json:"unit"`
	UnitPrice    float32   `json:"unit_price"`
	ReorderPoint float32   `json:"reorder_point"`
	LastUpdated  time.Time `json:"last_updated"`
}

type UpdateInventoryRequest struct {
	Name         *string  `json:"name"`
	Quantity     *float32 `json:"quantity"`
	Unit         *string  `json:"unit"`
	UnitPrice    *float32 `json:"unit_price"`
	ReorderPoint *float32 `json:"reorder_point"`
}

type CreateTransactionRequest struct {
	IngredientID    string  `json:"ingredient_id"`
	QuantityChange  float32 `json:"quantity_change"`
	TransactionType string  `json:"transaction_type"`
	Reason          string  `json:"reason"`
}

type TransactionResponse struct {
	TransactionID   string    `json:"transaction_id"`
	IngredientID    string    `json:"ingredient_id"`
	QuantityChange  float32   `json:"quantity_change"`
	TransactionType string    `json:"transaction_type"`
	Reason          string    `json:"reason"`
	CreatedAt       time.Time `json:"created_at"`
}
type LeftOverItem struct {
	Name     string  `json:"name"`
	Quantity float32 `json:"quantity"`
	Price    float32 `json:"price"`
}

type GetLeftOversResponse struct {
	CurrentPage int            `json:"currentPage"`
	HasNextPage bool           `json:"hasNextPage"`
	PageSize    int            `json:"pageSize"`
	TotalPages  int            `json:"totalPages"`
	Data        []LeftOverItem `json:"data"`
}

// type UpdateInventoryRequest struct {
// 	Name         string `json:"name"`
// 	StockLevel   int    `json:"stock_level"`
// 	UnitType     string `json:"unit_type"`
// 	ReOrderLevel int    `json:"re_order_level"`
// }
