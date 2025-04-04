package report

import (
	"context"
	"log"
	"strings"

	"frappuccino/internal/dto/report"
)

type SearchService struct {
	searchRepo searchRepo
	logger     *log.Logger
}

func NewSearchService(searchRepo searchRepo, logger *log.Logger) *SearchService {
	return &SearchService{
		searchRepo: searchRepo,
		logger:     logger,
	}
}

// Search performs a search operation based on parameters
func (s *SearchService) Search(ctx context.Context, req report.SearchRequest) (report.SearchResponse, error) {
	var response report.SearchResponse
	var minPrice, maxPrice *float64
	var err error

	// Prepare price filters
	if req.MinPrice > 0 {
		minPrice = &req.MinPrice
	}

	if req.MaxPrice > 0 {
		maxPrice = &req.MaxPrice
	}

	// Check if we should use keyword search
	useKeywords := false
	keywords := []string{}

	// Check if query contains words enclosed in quotes
	if strings.Contains(req.Query, "\"") {
		// Skip keyword search when using quoted phrases
		useKeywords = false
	} else {
		// Split query into individual keywords, filter empty words
		for _, word := range strings.Fields(req.Query) {
			if len(word) > 2 { // Skip very short words
				keywords = append(keywords, word)
			}
		}

		// Use keyword search if we have multiple keywords
		useKeywords = len(keywords) > 1
	}

	// Determine which entities to search based on filter parameter
	searchMenu := true
	searchOrders := true

	if req.Filter != "" {
		filters := strings.Split(req.Filter, ",")
		filterMap := make(map[string]bool)

		for _, f := range filters {
			filterMap[strings.TrimSpace(f)] = true
		}

		// Update search flags based on filter
		if !filterMap["all"] {
			searchMenu = filterMap["menu"]
			searchOrders = filterMap["orders"]
		}
	}

	// Search menu items if needed
	if searchMenu {
		var menuItems []report.SearchResultMenuItem

		if useKeywords {
			menuItems, err = s.searchRepo.SearchMenuItemsByKeywords(ctx, keywords, minPrice, maxPrice)
		} else {
			menuItems, err = s.searchRepo.SearchMenuItems(ctx, req.Query, minPrice, maxPrice)
		}

		if err != nil {
			s.logger.Printf("Error searching menu items: %v", err)
			// Continue with search instead of returning error
		} else {
			response.MenuItems = menuItems
			response.TotalMatches += len(menuItems)
		}
	}

	// Search orders if needed
	if searchOrders {
		var orders []report.SearchResultOrder

		if useKeywords {
			orders, err = s.searchRepo.SearchOrdersByKeywords(ctx, keywords, minPrice, maxPrice)
		} else {
			orders, err = s.searchRepo.SearchOrders(ctx, req.Query, minPrice, maxPrice)
		}

		if err != nil {
			s.logger.Printf("Error searching orders: %v", err)
			// Continue with search instead of returning error
		} else {
			response.Orders = orders
			response.TotalMatches += len(orders)
		}
	}

	return response, nil
}
