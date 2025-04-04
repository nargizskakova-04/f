package order

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	orderdto "frappuccino/internal/dto/order"
	"frappuccino/internal/entity"
	"frappuccino/internal/repository/postgres"
)

// processOrderWithTransaction processes a single order within a database transaction
func (s *OrderService) processOrderWithTransaction(ctx context.Context, req orderdto.CreateOrderRequest) (string, float64, error) {
	var items []entity.OrderItem
	var total float64

	// Begin a database transaction
	tx, err := s.orderRepo.Begin(ctx)
	if err != nil {
		return "", 0, fmt.Errorf("error starting transaction: %w", err)
	}

	// Ensure rollback on error
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				s.logger.Printf("Error rolling back transaction: %v", rbErr)
			}
		}
	}()

	// Get prices and build order items
	for _, dtoItem := range req.Items {
		// Get current price from the menu_items table
		price, err := s.orderRepo.GetMenuItemPrice(ctx, dtoItem.MenuItemID)
		if err != nil {
			return "", 0, fmt.Errorf("error getting price for item: %w", err)
		}

		// Calculate subtotal and accumulate total
		itemTotal := price * float64(dtoItem.Quantity)
		total += itemTotal

		// Ensure customizations is not nil
		customizations := dtoItem.Customizations
		if customizations == nil || len(customizations) == 0 {
			customizations = json.RawMessage(`{}`)
		}

		// Build order item entity
		items = append(items, entity.OrderItem{
			MenuItemID:     dtoItem.MenuItemID,
			Quantity:       dtoItem.Quantity,
			PriceAtTime:    price,
			Customizations: customizations,
		})
	}

	// Ensure special instructions is not nil
	specialInstructions := req.SpecialInstructions
	if specialInstructions == nil || len(specialInstructions) == 0 {
		specialInstructions = json.RawMessage(`{}`)
	}

	// Build order entity
	orderEntity := entity.Order{
		CustomerName:        req.CustomerName,
		SpecialInstructions: specialInstructions,
		TotalAmount:         total,
		Status:              "pending",
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	// Create the order and get the ID within the transaction
	orderID, err := s.orderRepo.CreateOrderWithTx(ctx, tx, orderEntity, items)
	if err != nil {
		return "", 0, fmt.Errorf("error creating order: %w", err)
	}

	// Deduct ingredients from inventory within the transaction
	err = s.deductIngredientsWithTransaction(ctx, tx, req.Items, orderID)
	if err != nil {
		return "", 0, fmt.Errorf("error deducting ingredients: %w", err)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return "", 0, fmt.Errorf("error committing transaction: %w", err)
	}

	return orderID, total, nil
}

// deductIngredientsWithTransaction updates inventory within a transaction
func (s *OrderService) deductIngredientsWithTransaction(
	ctx context.Context,
	tx *postgres.Transaction,
	items []orderdto.CreateOrderItem,
	orderID string,
) error {
	// Create a map to aggregate required quantities of ingredients
	requiredIngredients := make(map[string]float32)

	// For each menu item in the order
	for _, item := range items {
		// Get ingredients required for this menu item
		ingredients, err := s.menuRepo.GetMenuItemIngredients(ctx, item.MenuItemID)
		if err != nil {
			return fmt.Errorf("failed to get ingredients for menu item %s: %w", item.MenuItemID, err)
		}

		// For each ingredient, add the required quantity to our map
		for _, ing := range ingredients {
			// Multiply by the quantity of items ordered
			requiredQty := float32(ing.Quantity) * float32(item.Quantity)
			requiredIngredients[ing.IngredientID] += requiredQty
		}
	}

	// Update inventory for each ingredient
	for ingredientID, deductQty := range requiredIngredients {
		// Create inventory transaction record
		transaction := entity.InventoryTransaction{
			IngredientID:    ingredientID,
			QuantityChange:  deductQty,
			TransactionType: "deduction",
			Reason:          fmt.Sprintf("Order %s", orderID),
		}

		// Record the transaction within the transaction
		err := s.inventoryRepo.CreateInventoryTransactionWithTx(ctx, tx, transaction)
		if err != nil {
			return fmt.Errorf("failed to record transaction for ingredient %s: %w", ingredientID, err)
		}

		// Get current inventory
		inventory, err := s.inventoryRepo.GetInventoryByID(ctx, ingredientID)
		if err != nil {
			return fmt.Errorf("failed to get inventory for ingredient %s: %w", ingredientID, err)
		}

		// Check if we have enough inventory
		newQuantity := inventory.Quantity - deductQty
		if newQuantity < 0 {
			return fmt.Errorf("insufficient inventory for ingredient %s: required %.2f, available %.2f",
				ingredientID, deductQty, inventory.Quantity)
		}

		// Update inventory quantity within the transaction
		updates := map[string]interface{}{
			"quantity":     newQuantity,
			"last_updated": time.Now(),
		}

		err = s.inventoryRepo.UpdateInventoryWithTx(ctx, tx, updates, ingredientID)
		if err != nil {
			return fmt.Errorf("failed to update inventory for ingredient %s: %w", ingredientID, err)
		}

		// Check if we've fallen below reorder point
		if newQuantity <= inventory.ReorderPoint {
			s.logger.Printf("WARNING: Ingredient %s (%s) has fallen below reorder point. Current: %.2f, Reorder at: %.2f",
				inventory.Name, ingredientID, newQuantity, inventory.ReorderPoint)
			// In a real system, we might trigger a notification or automated order here
		}
	}

	return nil
}
