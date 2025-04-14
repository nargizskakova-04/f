package order

type BatchOrderRequest struct {
	Orders []CreateOrderRequest `json:"orders"`
}

type BatchOrderResult struct {
	OrderID      string  `json:"order_id,omitempty"`
	CustomerName string  `json:"customer_name"`
	Status       string  `json:"status"`
	Reason       string  `json:"reason,omitempty"`
	Total        float64 `json:"total,omitempty"`
}

type InventorySummary struct {
	IngredientID string  `json:"ingredient_id"`
	Name         string  `json:"name"`
	QuantityUsed float32 `json:"quantity_used"`
	Remaining    float32 `json:"remaining"`
}

type BatchOrderResponse struct {
	ProcessedOrders []BatchOrderResult `json:"processed_orders"`
	Summary         BatchSummary       `json:"summary"`
}

type BatchSummary struct {
	TotalOrders      int                `json:"total_orders"`
	Accepted         int                `json:"accepted"`
	Rejected         int                `json:"rejected"`
	TotalRevenue     float64            `json:"total_revenue"`
	InventoryUpdates []InventorySummary `json:"inventory_updates"`
}
