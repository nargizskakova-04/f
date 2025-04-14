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

type OrderService struct {
	orderRepo     orderRepo
	menuRepo      menuRepo
	inventoryRepo inventoryRepo
	logger        *log.Logger
}

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

type IngredientRequirement struct {
	IngredientID string
	Name         string
	Required     float32
	Available    float32
	Unit         string
}

func (s *OrderService) CreateOrder(ctx context.Context, req orderdto.CreateOrderRequest) (string, error) {
	var items []entity.OrderItem
	var total float64

	missingIngredients, err := s.validateIngredientsAvailability(ctx, req.Items)
	if err != nil {
		return "", fmt.Errorf("error validating ingredients: %w", err)
	}

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

	for _, dtoItem := range req.Items {
		price, err := s.orderRepo.GetMenuItemPrice(ctx, dtoItem.MenuItemID)
		if err != nil {
			s.logger.Println("Error getting price for item:", dtoItem.MenuItemID, err)
			return "", err
		}

		itemTotal := price * float64(dtoItem.Quantity)
		total += itemTotal

		customizations := dtoItem.Customizations
		if len(customizations) == 0 {
			customizations = json.RawMessage(`{}`)
		}

		items = append(items, entity.OrderItem{
			MenuItemID:     dtoItem.MenuItemID,
			Quantity:       dtoItem.Quantity,
			PriceAtTime:    price,
			Customizations: customizations,
		})
	}

	orderEntity := entity.Order{
		CustomerName:        req.CustomerName,
		SpecialInstructions: req.SpecialInstructions,
		TotalAmount:         total,
		Status:              "pending",
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	orderID, err := s.orderRepo.CreateOrder(ctx, orderEntity, items)
	if err != nil {
		s.logger.Println("Error creating order:", err)
		return "", err
	}

	err = s.deductIngredientsFromInventory(ctx, req.Items, orderID)
	if err != nil {
		s.logger.Println("Error deducting ingredients:", err)
		return "", fmt.Errorf("order created but failed to update inventory: %w", err)
	}

	return orderID, nil
}

func (s *OrderService) validateIngredientsAvailability(ctx context.Context, items []orderdto.CreateOrderItem) ([]IngredientRequirement, error) {
	requiredIngredients := make(map[string]float32)

	for _, item := range items {
		ingredients, err := s.menuRepo.GetMenuItemIngredients(ctx, item.MenuItemID)
		if err != nil {
			return nil, fmt.Errorf("failed to get ingredients for menu item %s: %w", item.MenuItemID, err)
		}

		for _, ing := range ingredients {
			requiredQty := float32(ing.Quantity) * float32(item.Quantity)
			requiredIngredients[ing.IngredientID] += requiredQty
		}
	}

	var missingIngredients []IngredientRequirement

	for ingredientID, requiredQty := range requiredIngredients {
		inventory, err := s.inventoryRepo.GetInventoryByID(ctx, ingredientID)
		if err != nil {
			return nil, fmt.Errorf("failed to get inventory for ingredient %s: %w", ingredientID, err)
		}

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

func (s *OrderService) deductIngredientsFromInventory(ctx context.Context, items []orderdto.CreateOrderItem, orderID string) error {
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

		err := s.inventoryRepo.CreateInventoryTransaction(ctx, transaction)
		if err != nil {
			return fmt.Errorf("failed to record transaction for ingredient %s: %w", ingredientID, err)
		}

		inventory, err := s.inventoryRepo.GetInventoryByID(ctx, ingredientID)
		if err != nil {
			return fmt.Errorf("failed to get inventory for ingredient %s: %w", ingredientID, err)
		}

		newQuantity := inventory.Quantity - deductQty
		updates := map[string]interface{}{
			"quantity":     newQuantity,
			"last_updated": time.Now(),
		}

		_, err = s.inventoryRepo.UpdateInventory(ctx, updates, ingredientID)
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
	itemCounts, err := s.orderRepo.GetNumberOfOrderedItems(ctx, startDate, endDate)
	if err != nil {
		s.logger.Printf("Error getting number of ordered items: %v", err)
		return nil, err
	}

	menuItems, err := s.menuRepo.GetMenuItem(ctx)
	if err != nil {
		s.logger.Printf("Error getting menu items: %v", err)
	} else {
		for _, item := range menuItems {
			if _, exists := itemCounts[item.Name]; !exists {
				itemCounts[item.Name] = 0
			}
		}
	}

	return itemCounts, nil
}
