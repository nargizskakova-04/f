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

func (s *OrderService) processOrderWithTransaction(ctx context.Context, req orderdto.CreateOrderRequest) (string, float64, error) {
	var items []entity.OrderItem
	var total float64

	tx, err := s.orderRepo.Begin(ctx)
	if err != nil {
		return "", 0, fmt.Errorf("error starting transaction: %w", err)
	}

	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				s.logger.Printf("Error rolling back transaction: %v", rbErr)
			}
		}
	}()

	for _, dtoItem := range req.Items {
		price, err := s.orderRepo.GetMenuItemPrice(ctx, dtoItem.MenuItemID)
		if err != nil {
			return "", 0, fmt.Errorf("error getting price for item: %w", err)
		}

		itemTotal := price * float64(dtoItem.Quantity)
		total += itemTotal

		customizations := dtoItem.Customizations
		if customizations == nil || len(customizations) == 0 {
			customizations = json.RawMessage(`{}`)
		}

		items = append(items, entity.OrderItem{
			MenuItemID:     dtoItem.MenuItemID,
			Quantity:       dtoItem.Quantity,
			PriceAtTime:    price,
			Customizations: customizations,
		})
	}

	specialInstructions := req.SpecialInstructions
	if specialInstructions == nil || len(specialInstructions) == 0 {
		specialInstructions = json.RawMessage(`{}`)
	}

	orderEntity := entity.Order{
		CustomerName:        req.CustomerName,
		SpecialInstructions: specialInstructions,
		TotalAmount:         total,
		Status:              "pending",
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	orderID, err := s.orderRepo.CreateOrderWithTx(ctx, tx, orderEntity, items)
	if err != nil {
		return "", 0, fmt.Errorf("error creating order: %w", err)
	}

	err = s.deductIngredientsWithTransaction(ctx, tx, req.Items, orderID)
	if err != nil {
		return "", 0, fmt.Errorf("error deducting ingredients: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return "", 0, fmt.Errorf("error committing transaction: %w", err)
	}

	return orderID, total, nil
}

func (s *OrderService) deductIngredientsWithTransaction(
	ctx context.Context,
	tx *postgres.Transaction,
	items []orderdto.CreateOrderItem,
	orderID string,
) error {
	requiredIngredients := make(map[string]float32)

	for _, item := range items {
		ingredients, err := s.menuRepo.GetMenuItemIngredients(ctx, item.MenuItemID)
		if err != nil {
			return fmt.Errorf("failed to get ingredients for menu item %s: %w", item.MenuItemID, err)
		}

		for _, ing := range ingredients {
			requiredQty := float32(ing.Quantity) * float32(item.Quantity)
			requiredIngredients[ing.IngredientID] += requiredQty
		}
	}

	for ingredientID, deductQty := range requiredIngredients {
		transaction := entity.InventoryTransaction{
			IngredientID:    ingredientID,
			QuantityChange:  deductQty,
			TransactionType: "deduction",
			Reason:          fmt.Sprintf("Order %s", orderID),
		}

		err := s.inventoryRepo.CreateInventoryTransactionWithTx(ctx, tx, transaction)
		if err != nil {
			return fmt.Errorf("failed to record transaction for ingredient %s: %w", ingredientID, err)
		}

		inventory, err := s.inventoryRepo.GetInventoryByID(ctx, ingredientID)
		if err != nil {
			return fmt.Errorf("failed to get inventory for ingredient %s: %w", ingredientID, err)
		}

		newQuantity := inventory.Quantity - deductQty
		if newQuantity < 0 {
			return fmt.Errorf("insufficient inventory for ingredient %s: required %.2f, available %.2f",
				ingredientID, deductQty, inventory.Quantity)
		}

		updates := map[string]interface{}{
			"quantity":     newQuantity,
			"last_updated": time.Now(),
		}

		err = s.inventoryRepo.UpdateInventoryWithTx(ctx, tx, updates, ingredientID)
		if err != nil {
			return fmt.Errorf("failed to update inventory for ingredient %s: %w", ingredientID, err)
		}

		if newQuantity <= inventory.ReorderPoint {
			s.logger.Printf("WARNING: Ingredient %s (%s) has fallen below reorder point. Current: %.2f, Reorder at: %.2f",
				inventory.Name, ingredientID, newQuantity, inventory.ReorderPoint)
		}
	}

	return nil
}
