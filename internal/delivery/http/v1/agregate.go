package v1

import (
	"encoding/json"
	"net/http"
	"strconv"

	"frappuccino/internal/dto/report"
)

// GetTotalSales handles the GET /reports/total-sales endpoint
func (h *ReportHandler) GetTotalSales(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	startDateStr := r.URL.Query().Get("startDate")
	endDateStr := r.URL.Query().Get("endDate")
	status := r.URL.Query().Get("status")

	// Initialize request
	req := report.TotalSalesRequest{
		Status: status,
	}

	// Parse dates if provided
	if startDateStr != "" {
		startDate, err := parseDate(startDateStr)
		if err != nil {
			h.logger.Printf("Invalid startDate format: %v", err)
			http.Error(w, "Invalid startDate format. Please use YYYY-MM-DD format.", http.StatusBadRequest)
			return
		}
		req.StartDate = &startDate
	}

	if endDateStr != "" {
		endDate, err := parseDate(endDateStr)
		if err != nil {
			h.logger.Printf("Invalid endDate format: %v", err)
			http.Error(w, "Invalid endDate format. Please use YYYY-MM-DD format.", http.StatusBadRequest)
			return
		}
		req.EndDate = &endDate
	}

	// Get total sales
	response, err := h.reportService.GetTotalSales(r.Context(), req)
	if err != nil {
		h.logger.Printf("Error getting total sales: %v", err)
		http.Error(w, "Error generating total sales report", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Printf("Error encoding total sales response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

// GetPopularItems handles the GET /reports/popular-items endpoint
func (h *ReportHandler) GetPopularItems(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	startDateStr := r.URL.Query().Get("startDate")
	endDateStr := r.URL.Query().Get("endDate")
	limitStr := r.URL.Query().Get("limit")

	// Initialize request
	req := report.PopularItemsRequest{}

	// Parse dates if provided
	if startDateStr != "" {
		startDate, err := parseDate(startDateStr)
		if err != nil {
			h.logger.Printf("Invalid startDate format: %v", err)
			http.Error(w, "Invalid startDate format. Please use YYYY-MM-DD format.", http.StatusBadRequest)
			return
		}
		req.StartDate = &startDate
	}

	if endDateStr != "" {
		endDate, err := parseDate(endDateStr)
		if err != nil {
			h.logger.Printf("Invalid endDate format: %v", err)
			http.Error(w, "Invalid endDate format. Please use YYYY-MM-DD format.", http.StatusBadRequest)
			return
		}
		req.EndDate = &endDate
	}

	// Parse limit if provided
	if limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			h.logger.Printf("Invalid limit parameter: %v", err)
			http.Error(w, "Invalid limit parameter. Must be a positive integer.", http.StatusBadRequest)
			return
		}
		req.Limit = limit
	}

	// Get popular items
	response, err := h.reportService.GetPopularItems(r.Context(), req)
	if err != nil {
		h.logger.Printf("Error getting popular items: %v", err)
		http.Error(w, "Error generating popular items report", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Printf("Error encoding popular items response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
