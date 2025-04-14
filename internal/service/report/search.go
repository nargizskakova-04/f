package report

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"frappuccino/internal/dto/report"
)

type SearchService struct {
	searchRepo searchRepo
	orderRepo  orderRepo
	logger     *log.Logger
}

func NewSearchService(
	searchRepo searchRepo,
	orderRepo orderRepo,
	logger *log.Logger,
) *SearchService {
	return &SearchService{
		searchRepo: searchRepo,
		orderRepo:  orderRepo,
		logger:     logger,
	}
}

func (s *SearchService) Search(ctx context.Context, req report.SearchRequest) (report.SearchResponse, error) {
	var response report.SearchResponse
	var minPrice, maxPrice *float64
	var err error

	if req.MinPrice > 0 {
		minPrice = &req.MinPrice
	}

	if req.MaxPrice > 0 {
		maxPrice = &req.MaxPrice
	}

	useKeywords := false
	keywords := []string{}

	if strings.Contains(req.Query, "\"") {
		useKeywords = false
	} else {
		for _, word := range strings.Fields(req.Query) {
			if len(word) > 2 {
				keywords = append(keywords, word)
			}
		}

		useKeywords = len(keywords) > 1
	}

	searchMenu := true
	searchOrders := true

	if req.Filter != "" {
		filters := strings.Split(req.Filter, ",")
		filterMap := make(map[string]bool)

		for _, f := range filters {
			filterMap[strings.TrimSpace(f)] = true
		}

		if !filterMap["all"] {
			searchMenu = filterMap["menu"]
			searchOrders = filterMap["orders"]
		}
	}

	if searchMenu {
		var menuItems []report.SearchResultMenuItem

		if useKeywords {
			menuItems, err = s.searchRepo.SearchMenuItemsByKeywords(ctx, keywords, minPrice, maxPrice)
		} else {
			menuItems, err = s.searchRepo.SearchMenuItems(ctx, req.Query, minPrice, maxPrice)
		}

		if err != nil {
			s.logger.Printf("Error searching menu items: %v", err)
		} else {
			response.MenuItems = menuItems
			response.TotalMatches += len(menuItems)
		}
	}

	if searchOrders {
		var orders []report.SearchResultOrder

		if useKeywords {
			orders, err = s.searchRepo.SearchOrdersByKeywords(ctx, keywords, minPrice, maxPrice)
		} else {
			orders, err = s.searchRepo.SearchOrders(ctx, req.Query, minPrice, maxPrice)
		}

		if err != nil {
			s.logger.Printf("Error searching orders: %v", err)
		} else {
			response.Orders = orders
			response.TotalMatches += len(orders)
		}
	}

	return response, nil
}

func (s *SearchService) GetOrderedItemsByPeriod(ctx context.Context, req report.OrderedItemsByPeriodRequest) (report.OrderedItemsByPeriodResponse, error) {
	var response report.OrderedItemsByPeriodResponse
	response.Period = req.Period

	var year int
	if req.Year == "" {
		year = time.Now().Year()
	} else {
		var err error
		year, err = strconv.Atoi(req.Year)
		if err != nil {
			return response, fmt.Errorf("invalid year format: %w", err)
		}
	}
	response.Year = strconv.Itoa(year)

	if req.Period == "day" {
		if req.Month == "" {
			return response, fmt.Errorf("month parameter is required when period is day")
		}

		monthMap := map[string]time.Month{
			"january":   time.January,
			"february":  time.February,
			"march":     time.March,
			"april":     time.April,
			"may":       time.May,
			"june":      time.June,
			"july":      time.July,
			"august":    time.August,
			"september": time.September,
			"october":   time.October,
			"november":  time.November,
			"december":  time.December,
		}

		month, ok := monthMap[strings.ToLower(req.Month)]
		if !ok {
			return response, fmt.Errorf("invalid month name: %s", req.Month)
		}

		response.Month = req.Month

		dayCounts, err := s.orderRepo.GetOrderedItemsByDay(ctx, month, year)
		if err != nil {
			s.logger.Printf("Error getting ordered items by day: %v", err)
			return response, err
		}

		response.OrderedItems = dayCounts

	} else if req.Period == "month" {
		monthCounts, err := s.orderRepo.GetOrderedItemsByMonth(ctx, year)
		if err != nil {
			s.logger.Printf("Error getting ordered items by month: %v", err)
			return response, err
		}

		response.OrderedItems = monthCounts

	} else {
		return response, fmt.Errorf("invalid period parameter: %s (must be 'day' or 'month')", req.Period)
	}

	return response, nil
}
