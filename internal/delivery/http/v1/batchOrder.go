package v1

import (
	"encoding/json"
	"net/http"

	orderdto "frappuccino/internal/dto/order"
)

func (h *OrderHandler) BatchProcessOrdersRequest(w http.ResponseWriter, r *http.Request) {
	var request orderdto.BatchOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Println("method:BatchProcessOrdersRequest, function:json decode", err.Error())
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	if len(request.Orders) == 0 {
		h.logger.Println("method:BatchProcessOrdersRequest, error:empty orders list")
		http.Error(w, "No orders provided", http.StatusBadRequest)
		return
	}

	response, err := h.orderService.BatchProcessOrders(r.Context(), request)
	if err != nil {
		h.logger.Println("method:BatchProcessOrdersRequest, function:BatchProcessOrders", err.Error())
		http.Error(w, "Error processing batch orders", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Println("method:BatchProcessOrdersRequest, function:json encode", err.Error())
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
