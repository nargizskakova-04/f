package v1

import (
	"encoding/json"
	"net/http"
	"strconv"

	"frappuccino/internal/dto/report"
)

// SearchReport handles the GET /reports/search endpoint
func (h *ReportHandler) SearchReport(w http.ResponseWriter, r *http.Request) {
	// Extract query parameters
	query := r.URL.Query().Get("q")
	if query == "" {
		h.logger.Println("Search query is required")
		http.Error(w, "Search query is required", http.StatusBadRequest)
		return
	}

	filter := r.URL.Query().Get("filter")
	minPriceStr := r.URL.Query().Get("minPrice")
	maxPriceStr := r.URL.Query().Get("maxPrice")

	// Initialize request with query
	req := report.SearchRequest{
		Query:  query,
		Filter: filter,
	}

	// Parse price parameters if provided
	if minPriceStr != "" {
		minPrice, err := strconv.ParseFloat(minPriceStr, 64)
		if err != nil {
			h.logger.Printf("Invalid minPrice parameter: %v", err)
			http.Error(w, "Invalid minPrice parameter", http.StatusBadRequest)
			return
		}
		req.MinPrice = minPrice
	}

	if maxPriceStr != "" {
		maxPrice, err := strconv.ParseFloat(maxPriceStr, 64)
		if err != nil {
			h.logger.Printf("Invalid maxPrice parameter: %v", err)
			http.Error(w, "Invalid maxPrice parameter", http.StatusBadRequest)
			return
		}
		req.MaxPrice = maxPrice
	}

	// Perform search
	response, err := h.reportService.Search(r.Context(), req)
	if err != nil {
		h.logger.Printf("Search error: %v", err)
		http.Error(w, "Error performing search", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Printf("Error encoding search response: %v", err)
		http.Error(w, "Error encoding search response", http.StatusInternalServerError)
		return
	}
}
