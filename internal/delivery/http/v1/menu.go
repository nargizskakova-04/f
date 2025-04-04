package v1

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"frappuccino/internal/dto/menu"
)

func (h *MenuHandler) CreateMenuItemRequest(w http.ResponseWriter, r *http.Request) {
	var request menu.CreateMenuItemRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Println("method:CreateMenuItemRequest, function:json decode", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := h.menuService.CreateMenuItem(r.Context(), request)
	if err != nil {
		h.logger.Println("method:CreateMenuItemRequest, function:CreateMenuItem", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(id); err != nil {
		h.logger.Println("method:CreateMenuItemRequest, function:json encode", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *MenuHandler) GetMenuResponse(w http.ResponseWriter, r *http.Request) {
	// For GET requests, we typically don't need to decode the request body
	// Instead, we directly call the service

	menuItems, err := h.menuService.GetMenuItem(r.Context())
	if err != nil {
		h.logger.Println("method:GetMenuItemRequest, function:GetMenuItem", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(menuItems); err != nil {
		h.logger.Println("method:GetMenuItemRequest, function:json encode", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *MenuHandler) GetMenuByIDResponse(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path (Go 1.22+ pattern matching)
	id := r.PathValue("id")
	if id == "" {
		h.logger.Println("method:GetMenuByIDRequest, function: missing id parameter")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Call service to get the menu item
	menuItem, err := h.menuService.GetMenuByID(r.Context(), id)
	if err != nil {
		h.logger.Println("method:GetMenuByIDRequest, function:GetMenuByID", err.Error())

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

	if err := json.NewEncoder(w).Encode(menuItem); err != nil {
		h.logger.Println("method:GetMenuByIDRequest, function:json encode", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *MenuHandler) DeleteMenuRequest(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		h.logger.Println("method:DeleteMenuRequest, function: missing id parameter")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ingredient_id, err := h.menuService.DeleteMenu(r.Context(), id)
	if err != nil {
		h.logger.Println("method:DeleteMenuRequest, function:DeleteMenu", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(ingredient_id); err != nil {
		h.logger.Println("method:DeleteMenuRequest, function:json encode", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *MenuHandler) UpdateMenuRequest(w http.ResponseWriter, r *http.Request) {
	var request menu.UpdateMenuRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Println("method:UpdateMenuRequest, function:json decode", err.Error())
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		h.logger.Println("method:UpdateMenuRequest, function: missing id parameter")
		http.Error(w, "Missing menu ID", http.StatusBadRequest)
		return
	}

	ingredient_id, err := h.menuService.UpdateMenu(r.Context(), request, id)
	if err != nil {
		h.logger.Println("method:UpdateMenuRequest, function:UpdateMenu", err.Error())
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
		h.logger.Println("method:UpdateMenuRequest, function:json encode", err.Error())
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

func (h *MenuHandler) GetAllPriceHistoryResponse(w http.ResponseWriter, r *http.Request) {
	// For GET requests, we typically don't need to decode the request body
	// Instead, we directly call the service

	priceHistory, err := h.menuService.GetAllPriceHistory(r.Context())
	if err != nil {
		h.logger.Println("method:GetAllPriceHistoryRequest, function:GetAllPriceHistory", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(priceHistory); err != nil {
		h.logger.Println("method:GetAllPriceHistoryRequest, function:json encode", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
