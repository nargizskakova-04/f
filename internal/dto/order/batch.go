package order

// BatchOrderRequest represents multiple orders to be processed in a single batch
type BatchOrderRequest struct {
	Orders []CreateOrderRequest `json:"orders"`
}

// BatchOrderResult represents the processing result for a single order
type BatchOrderResult struct {
	OrderID      string  `json:"order_id,omitempty"`
	CustomerName string  `json:"customer_name"`
	Status       string  `json:"status"`           // "accepted" or "rejected"
	Reason       string  `json:"reason,omitempty"` // reason for rejection if status is "rejected"
	Total        float64 `json:"total,omitempty"`
}

// InventorySummary represents inventory changes after batch processing
type InventorySummary struct {
	IngredientID string  `json:"ingredient_id"`
	Name         string  `json:"name"`
	QuantityUsed float32 `json:"quantity_used"`
	Remaining    float32 `json:"remaining"`
}

// BatchOrderResponse represents the response for batch order processing
type BatchOrderResponse struct {
	ProcessedOrders []BatchOrderResult `json:"processed_orders"`
	Summary         BatchSummary       `json:"summary"`
}

// BatchSummary contains aggregate information about the batch processing
type BatchSummary struct {
	TotalOrders      int                `json:"total_orders"`
	Accepted         int                `json:"accepted"`
	Rejected         int                `json:"rejected"`
	TotalRevenue     float64            `json:"total_revenue"`
	InventoryUpdates []InventorySummary `json:"inventory_updates"`
}
