package order

import (
	"context"
	"fmt"
	"sync"

	orderdto "frappuccino/internal/dto/order"
)

func (s *OrderService) BatchProcessOrders(ctx context.Context, req orderdto.BatchOrderRequest) (orderdto.BatchOrderResponse, error) {
	response := orderdto.BatchOrderResponse{
		ProcessedOrders: make([]orderdto.BatchOrderResult, len(req.Orders)),
		Summary: orderdto.BatchSummary{
			TotalOrders: len(req.Orders),
		},
	}

	var mutex sync.Mutex

	var wg sync.WaitGroup
	wg.Add(len(req.Orders))

	ingredientUsage := make(map[string]float32)
	ingredientNames := make(map[string]string)

	for _, order := range req.Orders {
		ingredients, err := s.calculateIngredientsNeeded(ctx, order.Items)
		if err != nil {
			continue
		}

		mutex.Lock()
		for id, usage := range ingredients {
			ingredientUsage[id] += usage.Required
			ingredientNames[id] = usage.Name
		}
		mutex.Unlock()
	}

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

	for i, order := range req.Orders {
		orderIndex := i
		orderRequest := order

		go func() {
			defer wg.Done()

			result := orderdto.BatchOrderResult{
				CustomerName: orderRequest.CustomerName,
			}

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

			result.Status = "accepted"
			result.OrderID = orderID
			result.Total = total

			mutex.Lock()
			response.ProcessedOrders[orderIndex] = result
			response.Summary.Accepted++
			response.Summary.TotalRevenue += total
			mutex.Unlock()
		}()
	}

	wg.Wait()

	for id, used := range ingredientUsage {
		if used == 0 {
			continue
		}

		inventory, err := s.inventoryRepo.GetInventoryByID(ctx, id)
		if err != nil {
			s.logger.Printf("Error getting inventory for summary: %v", err)
			continue
		}

		response.Summary.InventoryUpdates = append(response.Summary.InventoryUpdates, orderdto.InventorySummary{
			IngredientID: id,
			Name:         ingredientNames[id],
			QuantityUsed: used,
			Remaining:    inventory.Quantity,
		})
	}

	return response, nil
}

func (s *OrderService) calculateIngredientsNeeded(ctx context.Context, items []orderdto.CreateOrderItem) (map[string]IngredientRequirement, error) {
	requiredIngredients := make(map[string]IngredientRequirement)

	for _, item := range items {
		ingredients, err := s.menuRepo.GetMenuItemIngredients(ctx, item.MenuItemID)
		if err != nil {
			return nil, fmt.Errorf("failed to get ingredients for menu item %s: %w", item.MenuItemID, err)
		}

		for _, ing := range ingredients {
			requiredQty := float32(ing.Quantity) * float32(item.Quantity)

			req, exists := requiredIngredients[ing.IngredientID]
			if !exists {
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

			req.Required += requiredQty
			requiredIngredients[ing.IngredientID] = req
		}
	}

	return requiredIngredients, nil
}
