package v1

import (
	"encoding/json"
	"net/http"
	"strconv"

	"frappuccino/internal/dto/report"
)

func (h *ReportHandler) SearchReport(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		h.logger.Println("Search query is required")
		http.Error(w, "Search query is required", http.StatusBadRequest)
		return
	}

	filter := r.URL.Query().Get("filter")
	minPriceStr := r.URL.Query().Get("minPrice")
	maxPriceStr := r.URL.Query().Get("maxPrice")

	req := report.SearchRequest{
		Query:  query,
		Filter: filter,
	}

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

	response, err := h.reportService.Search(r.Context(), req)
	if err != nil {
		h.logger.Printf("Search error: %v", err)
		http.Error(w, "Error performing search", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Printf("Error encoding search response: %v", err)
		http.Error(w, "Error encoding search response", http.StatusInternalServerError)
		return
	}
}

func (h *ReportHandler) GetOrderedItemsByPeriod(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		h.logger.Println("Period parameter is required")
		http.Error(w, "Period parameter is required", http.StatusBadRequest)
		return
	}

	month := r.URL.Query().Get("month")
	year := r.URL.Query().Get("year")

	if period == "day" && month == "" {
		h.logger.Println("Month parameter is required when period is day")
		http.Error(w, "Month parameter is required when period is day", http.StatusBadRequest)
		return
	}

	req := report.OrderedItemsByPeriodRequest{
		Period: period,
		Month:  month,
		Year:   year,
	}

	response, err := h.reportService.GetOrderedItemsByPeriod(r.Context(), req)
	if err != nil {
		h.logger.Printf("Error getting ordered items by period: %v", err)
		http.Error(w, "Error retrieving ordered items data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Printf("Error encoding ordered items response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
