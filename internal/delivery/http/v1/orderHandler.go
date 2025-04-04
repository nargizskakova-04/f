package v1

import (
	"log"
	"net/http"
)

type OrderHandler struct {
	logger       *log.Logger
	orderService orderInterface
}

func NewOrderHandler(
	orderService orderInterface,
	logger *log.Logger,
) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
		logger:       logger,
	}
}

func SetOrderHandler(
	router *http.ServeMux,
	orderService orderInterface,
	logger *log.Logger,
) {
	handler := NewOrderHandler(orderService, logger)
	setOrderRoutes(handler, router)
}

// // OrderHandler is responsible for processing incoming HTTP requests related to orders.
// type OrderHandler struct {
// 	orderService order.ServiceInterface
// }

// // NewOrderHandler returns a new instance of OrderHandler with the given service injected.
// func NewOrderHandler(service order.ServiceInterface) *OrderHandler {
// 	return &OrderHandler{
// 		orderService: service,
// 	}
// }

// // CreateOrder handles the POST /orders endpoint.
// // It decodes the JSON request body, passes it to the service layer,
// // and sends a response containing the new order ID.
// func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
// 	var req orderdto.CreateOrderRequest

// 	// Step 1: Decode the incoming JSON request body into our DTO struct.
// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		http.Error(w, "Invalid request body", http.StatusBadRequest)
// 		return
// 	}

// 	// Step 2: Call the service layer to create a new order.
// 	orderID, err := h.orderService.CreateOrder(r.Context(), req)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	// Step 3: Prepare the success response payload.
// 	resp := map[string]string{
// 		"order_id": orderID,
// 		"status":   "created",
// 	}

// 	// Step 4: Send the JSON response with HTTP 201 Created status.
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusCreated)
// 	json.NewEncoder(w).Encode(resp)
// }
// func (h *OrderHandler) GetOrderByID(w http.ResponseWriter, r *http.Request) {
// 	id := strings.TrimPrefix(r.URL.Path, "/orders/")

// 	resp, err := h.orderService.GetOrderByID(r.Context(), id)
// 	if err != nil {
// 		http.Error(w, "Order not found", http.StatusNotFound)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(resp)
// }
