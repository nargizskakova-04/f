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
	orderRepo orderRepo
	logger    *log.Logger
}

func NewOrderService(orderRepo orderRepo, logger *log.Logger) *OrderService {
	return &OrderService{
		orderRepo: orderRepo,
		logger:    logger,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, req orderdto.CreateOrderRequest) (string, error) {
	var items []entity.OrderItem
	var total float64

	// Iterate over each item in the request
	for _, dtoItem := range req.Items {
		// Step 1: Get current price from the menu_items table
		price, err := s.orderRepo.GetMenuItemPrice(ctx, dtoItem.MenuItemID)
		if err != nil {
			s.logger.Println("Error getting price for item:", dtoItem.MenuItemID, err)
			return "", err
		}

		// Step 2: Calculate subtotal and accumulate total
		itemTotal := price * float64(dtoItem.Quantity)
		total += itemTotal
		defaultCustomizations := json.RawMessage(`{}`)
		// Step 3: Build order item entity
		items = append(items, entity.OrderItem{
			MenuItemID:     dtoItem.MenuItemID,
			Quantity:       dtoItem.Quantity,
			PriceAtTime:    price,
			Customizations: defaultCustomizations,
		})
	}

	// Step 4: Build order entity
	orderEntity := entity.Order{
		CustomerName:        req.CustomerName,
		SpecialInstructions: req.SpecialInstructions,
		TotalAmount:         total,
		Status:              "pending",
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	// Step 5: Call repository to insert order and items
	orderID, err := s.orderRepo.CreateOrder(ctx, orderEntity, items)
	if err != nil {
		s.logger.Println("Error creating order:", err)
		return "", err
	}

	return orderID, nil
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
		Items:               responseItems,
	}, nil
}

func (s *OrderService) GetAllOrders(ctx context.Context) ([]orderdto.GetOrderResponse, error) {
	// Call the repository function to get all orders
	orders, err := s.orderRepo.GetAllOrders(ctx)
	if err != nil {
		s.logger.Println("Error retrieving orders:", err)
		return nil, err
	}

	// Map entity.Order items to the response type
	var response []orderdto.GetOrderResponse
	for _, order := range orders {
		// Get order items for each order
		items, err := s.orderRepo.GetOrderItemsByOrderID(ctx, order.OrderID)
		if err != nil {
			s.logger.Println("Failed to get order items for order ID:", order.OrderID, err)
			return nil, err
		}

		// Map order items to response DTOs
		var responseItems []orderdto.GetOrderItemResponse
		for _, item := range items {
			responseItems = append(responseItems, orderdto.GetOrderItemResponse{
				MenuItemID:     item.MenuItemID,
				Quantity:       item.Quantity,
				PriceAtTime:    item.PriceAtTime,
				Customizations: item.Customizations,
			})
		}

		// Create response DTO for each order
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
	// Create a map for the updates
	updates := make(map[string]interface{})

	// Add fields to update map if they are provided
	if req.CustomerName != nil {
		updates["customer_name"] = *req.CustomerName
	}

	if req.SpecialInstructions != nil && len(req.SpecialInstructions) > 0 {
		updates["special_instructions"] = req.SpecialInstructions
	}

	if req.Status != nil {
		// Validate that the requested status is valid
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

		// Add change reason if status is updated
		if req.ChangeReason != nil {
			updates["change_reason"] = *req.ChangeReason
		} else {
			updates["change_reason"] = "" // Default empty reason
		}
	}

	// Only proceed if there are updates to make
	if len(updates) == 0 {
		return fmt.Errorf("no valid fields to update")
	}

	// Call repository to update order
	err := s.orderRepo.UpdateOrder(ctx, orderID, updates)
	if err != nil {
		s.logger.Printf("Error updating order: %v", err)
		return err
	}

	return nil
}

func (s *OrderService) GetAllOrderStatusHistory(ctx context.Context) ([]orderdto.OrderStatusHistoryResponse, error) {
	// Call repository to get all status history
	history, err := s.orderRepo.GetAllOrderStatusHistory(ctx)
	if err != nil {
		s.logger.Printf("Error getting all order status history: %v", err)
		return nil, err
	}

	// Map entity to DTO
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

// Add to OrderService
func (s *OrderService) CloseOrder(ctx context.Context, orderID string, reason string) error {
	// Create an update with status set to "delivered"
	updates := map[string]interface{}{
		"status": "delivered",
	}

	// Add change reason if provided
	if reason != "" {
		updates["change_reason"] = reason
	} else {
		updates["change_reason"] = "Order completed and delivered"
	}

	// Call repository to update order
	err := s.orderRepo.UpdateOrder(ctx, orderID, updates)
	if err != nil {
		s.logger.Printf("Error closing order: %v", err)
		return err
	}

	return nil
}
