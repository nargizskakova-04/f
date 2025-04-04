package v1

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"

	"frappuccino/internal/dto/order"
)

func (h *OrderHandler) CreateOrderRequest(w http.ResponseWriter, r *http.Request) {
	var request order.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Println("method:CreateOrderRequest, function:json decode", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := h.orderService.CreateOrder(r.Context(), request)
	if err != nil {
		h.logger.Println("method:CreateOrderRequest, function:CreateOrder", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(id); err != nil {
		h.logger.Println("method:CreateOrderRequest, function:json encode", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *OrderHandler) GetOrderByIDResponse(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path (Go 1.22+ pattern matching)
	id := r.PathValue("id")
	if id == "" {
		h.logger.Println("method:GetOrderByIDRequest, function: missing id parameter")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Call service to get the order item
	orderItem, err := h.orderService.GetOrderByID(r.Context(), id)
	if err != nil {
		h.logger.Println("method:GetOrderByIDRequest, function:GetOrderByID", err.Error())

		// Check if it's a "not found" error
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(orderItem); err != nil {
		h.logger.Println("method:GetOrderByIDRequest, function:json encode", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *OrderHandler) GetOrderResponse(w http.ResponseWriter, r *http.Request) {
	// For GET requests, we typically don't need to decode the request body
	// Instead, we directly call the service

	orderItems, err := h.orderService.GetAllOrders(r.Context())
	if err != nil {
		h.logger.Println("method:GetOrderItemRequest, function:GetOrderItem", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(orderItems); err != nil {
		h.logger.Println("method:GetOrderItemRequest, function:json encode", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *OrderHandler) UpdateOrderRequest(w http.ResponseWriter, r *http.Request) {
	var request order.UpdateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Println("method:UpdateOrderRequest, function:json decode", err.Error())
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		h.logger.Println("method:UpdateOrderRequest, function: missing id parameter")
		http.Error(w, "Missing order ID", http.StatusBadRequest)
		return
	}

	err := h.orderService.UpdateOrder(r.Context(), id, request)
	if err != nil {
		h.logger.Println("method:UpdateOrderRequest, function:UpdateOrder", err.Error())
		statusCode := http.StatusInternalServerError
		errorMessage := "Internal server error"

		// Check for specific errors to provide better responses
		if err.Error() == "no fields to update" {
			statusCode = http.StatusBadRequest
			errorMessage = "No fields to update"
		}

		http.Error(w, errorMessage, statusCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(id); err != nil {
		h.logger.Println("method:UpdateOrderRequest, function:json encode", err.Error())
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

// Add to your HTTP handlers
func (h *OrderHandler) GetAllOrderStatusHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	history, err := h.orderService.GetAllOrderStatusHistory(ctx)
	if err != nil {
		h.logger.Println("Error getting order status history:", err)
		http.Error(w, "Failed to get order status history", http.StatusInternalServerError)
		return
	}

	// Respond with JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(history); err != nil {
		h.logger.Println("Error encoding response:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *OrderHandler) DeleteOrderRequest(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		h.logger.Println("method:DeleteOrderRequest, function: missing id parameter")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ingredient_id, err := h.orderService.DeleteOrder(r.Context(), id)
	if err != nil {
		h.logger.Println("method:DeleteOrderRequest, function:DeleteOrder", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(ingredient_id); err != nil {
		h.logger.Println("method:DeleteOrderRequest, function:json encode", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *OrderHandler) CloseOrder(w http.ResponseWriter, r *http.Request) {
	// Extract order ID from URL path
	id := r.PathValue("id")
	if id == "" {
		h.logger.Println("method:CloseOrder, function: missing id parameter")
		http.Error(w, "Missing order ID", http.StatusBadRequest)
		return
	}

	// Decode request body
	var req order.CloseOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// If body is empty, continue with empty reason
		if err != io.EOF {
			h.logger.Println("method:CloseOrder, function:json decode", err.Error())
			http.Error(w, "Invalid request format", http.StatusBadRequest)
			return
		}
	}

	// Call service to close the order
	err := h.orderService.CloseOrder(r.Context(), id, req.Reason)
	if err != nil {
		h.logger.Println("method:CloseOrder, function:CloseOrder", err.Error())
		http.Error(w, "Failed to close order", http.StatusInternalServerError)
		return
	}

	// Success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]string{
		"message":  "Order closed successfully",
		"order_id": id,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Println("method:CloseOrder, function:json encode", err.Error())
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}
