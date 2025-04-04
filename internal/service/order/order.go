package order

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	orderdto "frappuccino/internal/dto/order"
	"frappuccino/internal/entity"
)

// OrderService handles business logic for orders
type OrderService struct {
	orderRepo     orderRepo
	menuRepo      menuRepo      // New dependency for accessing menu items and ingredients
	inventoryRepo inventoryRepo // New dependency for checking and updating inventory
	logger        *log.Logger
}

// NewOrderService creates a new order service with needed dependencies
func NewOrderService(
	orderRepo orderRepo,
	menuRepo menuRepo,
	inventoryRepo inventoryRepo,
	logger *log.Logger,
) *OrderService {
	return &OrderService{
		orderRepo:     orderRepo,
		menuRepo:      menuRepo,
		inventoryRepo: inventoryRepo,
		logger:        logger,
	}
}

// IngredientRequirement represents the amount of an ingredient needed
type IngredientRequirement struct {
	IngredientID string
	Name         string
	Required     float32
	Available    float32
	Unit         string
}

// CreateOrder handles the order creation with inventory validation
func (s *OrderService) CreateOrder(ctx context.Context, req orderdto.CreateOrderRequest) (string, error) {
	var items []entity.OrderItem
	var total float64

	// Step 1: Validate ingredients availability for all items in the order
	missingIngredients, err := s.validateIngredientsAvailability(ctx, req.Items)
	if err != nil {
		return "", fmt.Errorf("error validating ingredients: %w", err)
	}

	// If there are missing ingredients, return error with details
	if len(missingIngredients) > 0 {
		errorMsg := "Insufficient ingredients: "
		for i, ing := range missingIngredients {
			if i > 0 {
				errorMsg += ", "
			}
			errorMsg += fmt.Sprintf("%s (need %.2f %s, have %.2f %s)",
				ing.Name, ing.Required, ing.Unit, ing.Available, ing.Unit)
		}
		return "", fmt.Errorf(errorMsg)
	}

	// Step 2: Begin transaction - ideally this would be a database transaction
	// But we'll simulate it with our service logic

	// Get prices and build order items
	for _, dtoItem := range req.Items {
		// Get current price from the menu_items table
		price, err := s.orderRepo.GetMenuItemPrice(ctx, dtoItem.MenuItemID)
		if err != nil {
			s.logger.Println("Error getting price for item:", dtoItem.MenuItemID, err)
			return "", err
		}

		// Calculate subtotal and accumulate total
		itemTotal := price * float64(dtoItem.Quantity)
		total += itemTotal

		customizations := dtoItem.Customizations
		if len(customizations) == 0 {
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

	// Step 3: Build order entity
	orderEntity := entity.Order{
		CustomerName:        req.CustomerName,
		SpecialInstructions: req.SpecialInstructions,
		TotalAmount:         total,
		Status:              "pending",
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	// Step 4: Insert order and items
	orderID, err := s.orderRepo.CreateOrder(ctx, orderEntity, items)
	if err != nil {
		s.logger.Println("Error creating order:", err)
		return "", err
	}

	// Step 5: Deduct ingredients from inventory
	err = s.deductIngredientsFromInventory(ctx, req.Items, orderID)
	if err != nil {
		s.logger.Println("Error deducting ingredients:", err)
		// In a real system, we would rollback the order creation here
		// But for simplicity, we'll just log the error and continue
		return "", fmt.Errorf("order created but failed to update inventory: %w", err)
	}

	return orderID, nil
}

// validateIngredientsAvailability checks if all required ingredients are available
func (s *OrderService) validateIngredientsAvailability(ctx context.Context, items []orderdto.CreateOrderItem) ([]IngredientRequirement, error) {
	// Create a map to aggregate required quantities of ingredients
	requiredIngredients := make(map[string]float32)

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
			requiredIngredients[ing.IngredientID] += requiredQty
		}
	}

	// Check if we have enough of each ingredient
	var missingIngredients []IngredientRequirement

	// Get current inventory for all required ingredients
	for ingredientID, requiredQty := range requiredIngredients {
		inventory, err := s.inventoryRepo.GetInventoryByID(ctx, ingredientID)
		if err != nil {
			return nil, fmt.Errorf("failed to get inventory for ingredient %s: %w", ingredientID, err)
		}

		// If we don't have enough, add to missing ingredients
		if inventory.Quantity < requiredQty {
			missingIngredients = append(missingIngredients, IngredientRequirement{
				IngredientID: ingredientID,
				Name:         inventory.Name,
				Required:     requiredQty,
				Available:    inventory.Quantity,
				Unit:         inventory.Unit,
			})
		}
	}

	return missingIngredients, nil
}

// deductIngredientsFromInventory updates inventory after order creation
func (s *OrderService) deductIngredientsFromInventory(ctx context.Context, items []orderdto.CreateOrderItem, orderID string) error {
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

		// Record the transaction
		err := s.inventoryRepo.CreateInventoryTransaction(ctx, transaction)
		if err != nil {
			return fmt.Errorf("failed to record transaction for ingredient %s: %w", ingredientID, err)
		}

		// Get current inventory
		inventory, err := s.inventoryRepo.GetInventoryByID(ctx, ingredientID)
		if err != nil {
			return fmt.Errorf("failed to get inventory for ingredient %s: %w", ingredientID, err)
		}

		// Update inventory quantity
		newQuantity := inventory.Quantity - deductQty
		updates := map[string]interface{}{
			"quantity":     newQuantity,
			"last_updated": time.Now(),
		}

		_, err = s.inventoryRepo.UpdateInventory(ctx, updates, ingredientID)
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

// Existing methods like GetOrderByID, GetAllOrders, etc. remain unchanged...
// Including the other methods for completeness but without changes

func (s *OrderService) GetOrderByID(ctx context.Context, id string) (orderdto.GetOrderResponse, error) {
	orderEntity, err := s.orderRepo.GetOrderByID(ctx, id)
	if err != nil {
		s.logger.Println("Order not found:", err)
		return orderdto.GetOrderResponse{}, err
	}

	items, err := s.orderRepo.GetOrderItemsByOrderID(ctx, id)
	if err != nil {
		s.logger.Println("Failed to get order items:", err)
		return orderdto.GetOrderResponse{}, err
	}

	var responseItems []orderdto.GetOrderItemResponse
	for _, item := range items {
		responseItems = append(responseItems, orderdto.GetOrderItemResponse{
			MenuItemID:     item.MenuItemID,
			Quantity:       item.Quantity,
			PriceAtTime:    item.PriceAtTime,
			Customizations: item.Customizations,
		})
	}

	return orderdto.GetOrderResponse{
		OrderID:             orderEntity.OrderID,
		CustomerName:        orderEntity.CustomerName,
		SpecialInstructions: orderEntity.SpecialInstructions,
		TotalAmount:         orderEntity.TotalAmount,
		Status:              orderEntity.Status,
		CreatedAt:           orderEntity.CreatedAt,
		UpdatedAt:           orderEntity.UpdatedAt,
		Items:               responseItems,
	}, nil
}

func (s *OrderService) GetAllOrders(ctx context.Context) ([]orderdto.GetOrderResponse, error) {
	orders, err := s.orderRepo.GetAllOrders(ctx)
	if err != nil {
		s.logger.Println("Error retrieving orders:", err)
		return nil, err
	}

	var response []orderdto.GetOrderResponse
	for _, order := range orders {
		items, err := s.orderRepo.GetOrderItemsByOrderID(ctx, order.OrderID)
		if err != nil {
			s.logger.Println("Failed to get order items for order ID:", order.OrderID, err)
			return nil, err
		}

		var responseItems []orderdto.GetOrderItemResponse
		for _, item := range items {
			responseItems = append(responseItems, orderdto.GetOrderItemResponse{
				MenuItemID:     item.MenuItemID,
				Quantity:       item.Quantity,
				PriceAtTime:    item.PriceAtTime,
				Customizations: item.Customizations,
			})
		}

		response = append(response, orderdto.GetOrderResponse{
			OrderID:             order.OrderID,
			CustomerName:        order.CustomerName,
			SpecialInstructions: order.SpecialInstructions,
			TotalAmount:         order.TotalAmount,
			Status:              order.Status,
			CreatedAt:           order.CreatedAt,
			UpdatedAt:           order.UpdatedAt,
			Items:               responseItems,
		})
	}

	return response, nil
}

func (s *OrderService) UpdateOrder(ctx context.Context, orderID string, req orderdto.UpdateOrderRequest) error {
	updates := make(map[string]interface{})

	if req.CustomerName != nil {
		updates["customer_name"] = *req.CustomerName
	}

	if req.SpecialInstructions != nil && len(req.SpecialInstructions) > 0 {
		updates["special_instructions"] = req.SpecialInstructions
	}

	if req.Status != nil {
		validStatuses := map[string]bool{
			"pending":   true,
			"preparing": true,
			"ready":     true,
			"delivered": true,
			"cancelled": true,
		}

		if !validStatuses[*req.Status] {
			s.logger.Printf("Invalid order status: %s", *req.Status)
			return fmt.Errorf("invalid order status: %s", *req.Status)
		}

		updates["status"] = *req.Status

		if req.ChangeReason != nil {
			updates["change_reason"] = *req.ChangeReason
		} else {
			updates["change_reason"] = ""
		}
	}

	if len(updates) == 0 {
		return fmt.Errorf("no valid fields to update")
	}

	err := s.orderRepo.UpdateOrder(ctx, orderID, updates)
	if err != nil {
		s.logger.Printf("Error updating order: %v", err)
		return err
	}

	return nil
}

func (s *OrderService) GetAllOrderStatusHistory(ctx context.Context) ([]orderdto.OrderStatusHistoryResponse, error) {
	history, err := s.orderRepo.GetAllOrderStatusHistory(ctx)
	if err != nil {
		s.logger.Printf("Error getting all order status history: %v", err)
		return nil, err
	}

	var response []orderdto.OrderStatusHistoryResponse
	for _, h := range history {
		response = append(response, orderdto.OrderStatusHistoryResponse{
			OrderStatusID: h.OrderStatusID,
			OrderID:       h.OrderID,
			OldStatus:     h.OldStatus,
			NewStatus:     h.NewStatus,
			ChangedAt:     h.ChangedAt,
			ChangeReason:  h.ChangeReason,
		})
	}

	return response, nil
}

func (s *OrderService) DeleteOrder(ctx context.Context, id string) (string, error) {
	order_id, err := s.orderRepo.DeleteOrder(ctx, id)
	if err != nil {
		s.logger.Println(err)
		return "", err
	}
	return order_id, nil
}

func (s *OrderService) CloseOrder(ctx context.Context, orderID string, reason string) error {
	updates := map[string]interface{}{
		"status": "delivered",
	}

	if reason != "" {
		updates["change_reason"] = reason
	} else {
		updates["change_reason"] = "Order completed and delivered"
	}

	err := s.orderRepo.UpdateOrder(ctx, orderID, updates)
	if err != nil {
		s.logger.Printf("Error closing order: %v", err)
		return err
	}

	return nil
}

func (s *OrderService) GetNumberOfOrderedItems(ctx context.Context, startDate, endDate *time.Time) (map[string]int, error) {
	// Call the repository to get the data
	itemCounts, err := s.orderRepo.GetNumberOfOrderedItems(ctx, startDate, endDate)
	if err != nil {
		s.logger.Printf("Error getting number of ordered items: %v", err)
		return nil, err
	}

	// Get the list of all menu items to ensure all items are represented in the response
	menuItems, err := s.menuRepo.GetMenuItem(ctx)
	if err != nil {
		s.logger.Printf("Error getting menu items: %v", err)
		// Not returning an error here, as we still have the order counts
		// We'll just skip adding the zero counts for menu items that weren't ordered
	} else {
		// Ensure all menu items are in the result map with at least 0 count
		for _, item := range menuItems {
			if _, exists := itemCounts[item.Name]; !exists {
				itemCounts[item.Name] = 0
			}
		}
	}

	return itemCounts, nil
}
