package v1

import (
	"encoding/json"
	"net/http"
	"strconv"

	"frappuccino/internal/dto/report"
)

func (h *ReportHandler) GetTotalSales(w http.ResponseWriter, r *http.Request) {
	startDateStr := r.URL.Query().Get("startDate")
	endDateStr := r.URL.Query().Get("endDate")
	status := r.URL.Query().Get("status")

	req := report.TotalSalesRequest{
		Status: status,
	}

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

	response, err := h.reportService.GetTotalSales(r.Context(), req)
	if err != nil {
		h.logger.Printf("Error getting total sales: %v", err)
		http.Error(w, "Error generating total sales report", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Printf("Error encoding total sales response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func (h *ReportHandler) GetPopularItems(w http.ResponseWriter, r *http.Request) {
	startDateStr := r.URL.Query().Get("startDate")
	endDateStr := r.URL.Query().Get("endDate")
	limitStr := r.URL.Query().Get("limit")

	req := report.PopularItemsRequest{}

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

	if limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			h.logger.Printf("Invalid limit parameter: %v", err)
			http.Error(w, "Invalid limit parameter. Must be a positive integer.", http.StatusBadRequest)
			return
		}
		req.Limit = limit
	}

	response, err := h.reportService.GetPopularItems(r.Context(), req)
	if err != nil {
		h.logger.Printf("Error getting popular items: %v", err)
		http.Error(w, "Error generating popular items report", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Printf("Error encoding popular items response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
