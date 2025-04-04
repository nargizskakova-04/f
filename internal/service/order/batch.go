package order

import (
	"context"
	"fmt"
	"sync"

	orderdto "frappuccino/internal/dto/order"
)

// BatchProcessOrders processes multiple orders concurrently with inventory consistency
func (s *OrderService) BatchProcessOrders(ctx context.Context, req orderdto.BatchOrderRequest) (orderdto.BatchOrderResponse, error) {
	// Initialize response
	response := orderdto.BatchOrderResponse{
		ProcessedOrders: make([]orderdto.BatchOrderResult, len(req.Orders)),
		Summary: orderdto.BatchSummary{
			TotalOrders: len(req.Orders),
		},
	}

	// Create a mutex for synchronized access to shared resources
	var mutex sync.Mutex

	// Create a wait group to wait for all goroutines to finish
	var wg sync.WaitGroup
	wg.Add(len(req.Orders))

	// Store ingredient usage across all orders
	ingredientUsage := make(map[string]float32)
	ingredientNames := make(map[string]string)

	// First phase: Validate all orders and calculate total ingredient requirements
	// This pre-check helps avoid deadlocks and ensures we have enough inventory
	for _, order := range req.Orders {
		// For each order, check ingredient requirements
		ingredients, err := s.calculateIngredientsNeeded(ctx, order.Items)
		if err != nil {
			// If we can't calculate ingredients, we'll reject the order in the processing phase
			continue
		}

		// Add to total ingredient usage
		mutex.Lock()
		for id, usage := range ingredients {
			ingredientUsage[id] += usage.Required
			ingredientNames[id] = usage.Name
		}
		mutex.Unlock()
	}

	// Check if we have enough inventory for the entire batch
	insufficientIngredients := make(map[string]float32)
	for id, required := range ingredientUsage {
		inventory, err := s.inventoryRepo.GetInventoryByID(ctx, id)
		if err != nil {
			return response, fmt.Errorf("failed to get inventory for ingredient %s: %w", id, err)
		}

		if inventory.Quantity < required {
			insufficientIngredients[id] = inventory.Quantity
		}
	}

	// Process each order concurrently
	for i, order := range req.Orders {
		// Capture iteration variables for goroutine
		orderIndex := i
		orderRequest := order

		go func() {
			defer wg.Done()

			// Initialize result with customer name
			result := orderdto.BatchOrderResult{
				CustomerName: orderRequest.CustomerName,
			}

			// Check if we have enough inventory based on pre-check
			orderIngredients, err := s.calculateIngredientsNeeded(ctx, orderRequest.Items)
			if err != nil {
				result.Status = "rejected"
				result.Reason = "error_processing_order"

				mutex.Lock()
				response.ProcessedOrders[orderIndex] = result
				response.Summary.Rejected++
				mutex.Unlock()
				return
			}

			// Check if any ingredients are insufficient in the pre-check
			var insufficientFound bool
			for id, requirement := range orderIngredients {
				if remaining, exists := insufficientIngredients[id]; exists {
					result.Status = "rejected"
					result.Reason = fmt.Sprintf("insufficient_inventory: %s (need %.2f %s, have %.2f %s)",
						ingredientNames[id],
						requirement.Required,
						requirement.Unit,
						remaining,
						requirement.Unit)
					insufficientFound = true
					break
				}
			}

			if insufficientFound {
				mutex.Lock()
				response.ProcessedOrders[orderIndex] = result
				response.Summary.Rejected++
				mutex.Unlock()
				return
			}

			// Process the order with a transaction
			orderID, total, err := s.processOrderWithTransaction(ctx, orderRequest)
			if err != nil {
				result.Status = "rejected"
				result.Reason = fmt.Sprintf("error: %s", err.Error())

				mutex.Lock()
				response.ProcessedOrders[orderIndex] = result
				response.Summary.Rejected++
				mutex.Unlock()
				return
			}

			// Order successfully processed
			result.Status = "accepted"
			result.OrderID = orderID
			result.Total = total

			// Update response and summary atomically
			mutex.Lock()
			response.ProcessedOrders[orderIndex] = result
			response.Summary.Accepted++
			response.Summary.TotalRevenue += total
			mutex.Unlock()
		}()
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// After all orders are processed, collect inventory summary
	for id, used := range ingredientUsage {
		// Skip ingredients that weren't used successfully
		if used == 0 {
			continue
		}

		// Get current inventory state
		inventory, err := s.inventoryRepo.GetInventoryByID(ctx, id)
		if err != nil {
			s.logger.Printf("Error getting inventory for summary: %v", err)
			continue
		}

		// Add to summary
		response.Summary.InventoryUpdates = append(response.Summary.InventoryUpdates, orderdto.InventorySummary{
			IngredientID: id,
			Name:         ingredientNames[id],
			QuantityUsed: used,
			Remaining:    inventory.Quantity,
		})
	}

	return response, nil
}

// calculateIngredientsNeeded calculates the required ingredients for given order items
func (s *OrderService) calculateIngredientsNeeded(ctx context.Context, items []orderdto.CreateOrderItem) (map[string]IngredientRequirement, error) {
	requiredIngredients := make(map[string]IngredientRequirement)

	// For each menu item in the order
	for _, item := range items {
		// Get ingredients required for this menu item
		ingredients, err := s.menuRepo.GetMenuItemIngredients(ctx, item.MenuItemID)
		if err != nil {
			return nil, fmt.Errorf("failed to get ingredients for menu item %s: %w", item.MenuItemID, err)
		}

		// For each ingredient, add the required quantity to our map
		for _, ing := range ingredients {
			// Multiply by the quantity of items ordered
			requiredQty := float32(ing.Quantity) * float32(item.Quantity)

			// Get or initialize the requirement
			req, exists := requiredIngredients[ing.IngredientID]
			if !exists {
				// Get inventory to get the name and available quantity
				inventory, err := s.inventoryRepo.GetInventoryByID(ctx, ing.IngredientID)
				if err != nil {
					return nil, fmt.Errorf("failed to get inventory for ingredient %s: %w", ing.IngredientID, err)
				}

				req = IngredientRequirement{
					IngredientID: ing.IngredientID,
					Name:         inventory.Name,
					Required:     0,
					Available:    inventory.Quantity,
					Unit:         inventory.Unit,
				}
			}

			// Add the required quantity
			req.Required += requiredQty
			requiredIngredients[ing.IngredientID] = req
		}
	}

	return requiredIngredients, nil
}

// Reference to processOrderWithTransaction
// The actual implementation is in batch_tx.go
