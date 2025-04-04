package entity

import "time"

type Inventory struct {
	IngredientID string
	Name         string
	Quantity     float32
	Unit         string
	UnitPrice    float32
	LastUpdated  time.Time
	ReorderPoint float32
}

type InventoryTransaction struct {
	TransactionID   string
	IngredientID    string
	QuantityChange  float32
	TransactionType string
	Reason          string
	CreatedAt       time.Time
}
