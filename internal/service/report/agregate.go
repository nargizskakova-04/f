package report

import (
	"context"
	"time"

	"frappuccino/internal/dto/report"
)

func (s *SearchService) GetTotalSales(ctx context.Context, req report.TotalSalesRequest) (report.TotalSalesResponse, error) {
	totalSales, orderCount, err := s.orderRepo.GetTotalSales(ctx, req.StartDate, req.EndDate, req.Status)
	if err != nil {
		s.logger.Printf("Error getting total sales: %v", err)
		return report.TotalSalesResponse{}, err
	}

	var averageOrderSize float64
	if orderCount > 0 {
		averageOrderSize = totalSales / float64(orderCount)
	}

	response := report.TotalSalesResponse{
		TotalSales:       totalSales,
		OrderCount:       orderCount,
		AverageOrderSize: averageOrderSize,
		Status:           req.Status,
	}

	if req.StartDate != nil {
		response.StartDate = *req.StartDate
	} else {
		defaultStartDate := time.Now().AddDate(0, 0, -30)
		response.StartDate = defaultStartDate
	}

	if req.EndDate != nil {
		response.EndDate = *req.EndDate
	} else {
		response.EndDate = time.Now()
	}

	return response, nil
}

func (s *SearchService) GetPopularItems(ctx context.Context, req report.PopularItemsRequest) (report.PopularItemsResponse, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}

	items, totalQuantity, totalRevenue, err := s.orderRepo.GetPopularItems(ctx, req.StartDate, req.EndDate, limit)
	if err != nil {
		s.logger.Printf("Error getting popular items: %v", err)
		return report.PopularItemsResponse{}, err
	}

	response := report.PopularItemsResponse{
		Items:         items,
		TotalQuantity: totalQuantity,
		TotalRevenue:  totalRevenue,
	}

	if req.StartDate != nil {
		response.StartDate = *req.StartDate
	} else {
		defaultStartDate := time.Now().AddDate(0, 0, -30)
		response.StartDate = defaultStartDate
	}

	if req.EndDate != nil {
		response.EndDate = *req.EndDate
	} else {
		response.EndDate = time.Now()
	}

	if len(items) > 0 {
		mostPopular := items[0].Name

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
