package report

import (
	"context"
	"time"

	"frappuccino/internal/dto/report"
)

// GetTotalSales calculates the total sales for the given date range and status
func (s *SearchService) GetTotalSales(ctx context.Context, req report.TotalSalesRequest) (report.TotalSalesResponse, error) {
	// Get total sales from repository
	totalSales, orderCount, err := s.orderRepo.GetTotalSales(ctx, req.StartDate, req.EndDate, req.Status)
	if err != nil {
		s.logger.Printf("Error getting total sales: %v", err)
		return report.TotalSalesResponse{}, err
	}

	// Calculate average order size
	var averageOrderSize float64
	if orderCount > 0 {
		averageOrderSize = totalSales / float64(orderCount)
	}

	// Prepare response
	response := report.TotalSalesResponse{
		TotalSales:       totalSales,
		OrderCount:       orderCount,
		AverageOrderSize: averageOrderSize,
		Status:           req.Status,
	}

	// Set date range in the response
	if req.StartDate != nil {
		response.StartDate = *req.StartDate
	} else {
		// Use a default start date (e.g., 30 days ago) if not provided
		defaultStartDate := time.Now().AddDate(0, 0, -30)
		response.StartDate = defaultStartDate
	}

	if req.EndDate != nil {
		response.EndDate = *req.EndDate
	} else {
		// Use current time as default end date if not provided
		response.EndDate = time.Now()
	}

	return response, nil
}

// GetPopularItems retrieves the most popular menu items for the given date range
func (s *SearchService) GetPopularItems(ctx context.Context, req report.PopularItemsRequest) (report.PopularItemsResponse, error) {
	// Set default limit if not specified
	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}

	// Get popular items from repository
	items, totalQuantity, totalRevenue, err := s.orderRepo.GetPopularItems(ctx, req.StartDate, req.EndDate, limit)
	if err != nil {
		s.logger.Printf("Error getting popular items: %v", err)
		return report.PopularItemsResponse{}, err
	}

	// Prepare response
	response := report.PopularItemsResponse{
		Items:         items,
		TotalQuantity: totalQuantity,
		TotalRevenue:  totalRevenue,
	}

	// Set date range in the response
	if req.StartDate != nil {
		response.StartDate = *req.StartDate
	} else {
		// Use a default start date (e.g., 30 days ago) if not provided
		defaultStartDate := time.Now().AddDate(0, 0, -30)
		response.StartDate = defaultStartDate
	}

	if req.EndDate != nil {
		response.EndDate = *req.EndDate
	} else {
		// Use current time as default end date if not provided
		response.EndDate = time.Now()
	}

	// Set most popular and most revenue items
	if len(items) > 0 {
		mostPopular := items[0].Name // First item is the most popular (sorted by quantity)

		// Find item with highest revenue
		var mostRevenueItem report.PopularItem
		for _, item := range items {
			if item.TotalRevenue > mostRevenueItem.TotalRevenue {
				mostRevenueItem = item
			}
		}

		response.MostPopular = mostPopular
		response.MostRevenue = mostRevenueItem.Name
	} else {
		response.MostPopular = "No data available"
		response.MostRevenue = "No data available"
	}

	return response, nil
}
