package v1

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"frappuccino/internal/dto/inventory"
)

func (h *InventoryHandler) CreateInventoryRequest(w http.ResponseWriter, r *http.Request) {
	var request inventory.CreateInventoryRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Println("method:CreateInventoryRequest, function:json decode", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	id, err := h.inventoryService.CreateInventory(r.Context(), request)
	if err != nil {
		h.logger.Println("method:CreateInventoryRequest, function:CreateInventory", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(id); err != nil {
		h.logger.Println("method:CreateInventoryRequest, function:json encode", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *InventoryHandler) GetInventoryResponse(w http.ResponseWriter, r *http.Request) {
	// For GET requests, we typically don't need to decode the request body
	// Instead, we directly call the service

	inventoryItems, err := h.inventoryService.GetInventory(r.Context())
	if err != nil {
		h.logger.Println("method:GetInventoryRequest, function:GetInventory", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(inventoryItems); err != nil {
		h.logger.Println("method:GetInventoryRequest, function:json encode", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *InventoryHandler) GetInventoryByIDResponse(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path (Go 1.22+ pattern matching)
	id := r.PathValue("id")
	if id == "" {
		h.logger.Println("method:GetInventoryByIDRequest, function: missing id parameter")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Call service to get the inventory item
	inventoryItem, err := h.inventoryService.GetInventoryByID(r.Context(), id)
	if err != nil {
		h.logger.Println("method:GetInventoryByIDRequest, function:GetInventoryByID", err.Error())

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

	if err := json.NewEncoder(w).Encode(inventoryItem); err != nil {
		h.logger.Println("method:GetInventoryByIDRequest, function:json encode", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *InventoryHandler) DeleteInventoryRequest(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		h.logger.Println("method:DeleteInventoryRequest, function: missing id parameter")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ingredient_id, err := h.inventoryService.DeleteInventory(r.Context(), id)
	if err != nil {
		h.logger.Println("method:DeleteInventoryRequest, function:DeleteInventory", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(ingredient_id); err != nil {
		h.logger.Println("method:DeleteInventoryRequest, function:json encode", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *InventoryHandler) UpdateInventoryRequest(w http.ResponseWriter, r *http.Request) {
	var request inventory.UpdateInventoryRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Println("method:UpdateInventoryRequest, function:json decode", err.Error())
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		h.logger.Println("method:UpdateInventoryRequest, function: missing id parameter")
		http.Error(w, "Missing inventory ID", http.StatusBadRequest)
		return
	}

	ingredient_id, err := h.inventoryService.UpdateInventory(r.Context(), request, id)
	if err != nil {
		h.logger.Println("method:UpdateInventoryRequest, function:UpdateInventory", err.Error())
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
	if err := json.NewEncoder(w).Encode(ingredient_id); err != nil {
		h.logger.Println("method:UpdateInventoryRequest, function:json encode", err.Error())
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

func (h *InventoryHandler) CreateInventoryTransactionRequest(w http.ResponseWriter, r *http.Request) {
	var request inventory.CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Println("method:CreateInventoryTransactionRequest, function:json decode", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Validate transaction type
	validTypes := map[string]bool{
		"addition":   true,
		"deduction":  true,
		"adjustment": true,
		"waste":      true,
	}

	if !validTypes[request.TransactionType] {
		h.logger.Println("method:CreateInventoryTransactionRequest, invalid transaction type:", request.TransactionType)
		http.Error(w, "Invalid transaction type", http.StatusBadRequest)
		return
	}

	if request.QuantityChange <= 0 {
		h.logger.Println("method:CreateInventoryTransactionRequest, quantity must be positive")
		http.Error(w, "Quantity change must be a positive value", http.StatusBadRequest)
		return
	}

	err := h.inventoryService.RecordInventoryTransaction(r.Context(), request)
	if err != nil {
		h.logger.Println("method:CreateInventoryTransactionRequest, function:RecordInventoryTransaction", err.Error())

		if err.Error() == "ingredient not found" {
			http.Error(w, "Ingredient not found", http.StatusNotFound)
			return
		}

		if err.Error() == "insufficient inventory" {
			http.Error(w, "Insufficient inventory", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]string{"message": "Transaction recorded successfully"}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Println("method:CreateInventoryTransactionRequest, function:json encode", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (h *InventoryHandler) GetInventoryTransactionsResponse(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		h.logger.Println("method:GetInventoryTransactionsResponse, function: missing id parameter")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	transactions, err := h.inventoryService.GetInventoryTransactions(r.Context(), id)
	if err != nil {
		h.logger.Println("method:GetInventoryTransactionsResponse, function:GetInventoryTransactions", err.Error())

		if err.Error() == "ingredient not found" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(transactions); err != nil {
		h.logger.Println("method:GetInventoryTransactionsResponse, function:json encode", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *InventoryHandler) GetLeftOversResponse(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()
	sortBy := query.Get("sortBy")
	pageStr := query.Get("page")
	pageSizeStr := query.Get("pageSize")

	// Set defaults and parse page parameters
	page := 1
	pageSize := 10

	if pageStr != "" {
		parsedPage, err := strconv.Atoi(pageStr)
		if err != nil || parsedPage < 1 {
			h.logger.Println("method:GetLeftOversResponse, invalid page parameter:", pageStr)
			http.Error(w, "Invalid page parameter, must be a positive integer", http.StatusBadRequest)
			return
		}
		page = parsedPage
	}

	if pageSizeStr != "" {
		parsedPageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil || parsedPageSize < 1 {
			h.logger.Println("method:GetLeftOversResponse, invalid pageSize parameter:", pageSizeStr)
			http.Error(w, "Invalid pageSize parameter, must be a positive integer", http.StatusBadRequest)
			return
		}
		pageSize = parsedPageSize
	}

	// Clean and validate sortBy parameter
	// The issue here is that the URL might have a format like "sortBy=quantity?page=1"
	// where the query string contains another '?' instead of '&'
	if strings.Contains(sortBy, "?") {
		// Extract the actual sortBy value before any additional '?'
		sortBy = strings.Split(sortBy, "?")[0]
	}

	// Now validate the clean sortBy parameter
	if sortBy != "" && sortBy != "price" && sortBy != "quantity" {
		h.logger.Println("method:GetLeftOversResponse, invalid sortBy parameter:", sortBy)
		http.Error(w, "Invalid sortBy parameter. Must be 'price' or 'quantity'", http.StatusBadRequest)
		return
	}

	// Call service to get leftovers
	response, err := h.inventoryService.GetLeftOvers(r.Context(), sortBy, page, pageSize)
	if err != nil {
		h.logger.Println("method:GetLeftOversResponse, function:GetLeftOvers", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Println("method:GetLeftOversResponse, function:json encode", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
