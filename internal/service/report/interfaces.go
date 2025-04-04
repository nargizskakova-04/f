package report

import (
	"context"

	"frappuccino/internal/dto/report"
)

type searchRepo interface {
	SearchMenuItems(ctx context.Context, query string, minPrice, maxPrice *float64) ([]report.SearchResultMenuItem, error)
	SearchOrders(ctx context.Context, query string, minPrice, maxPrice *float64) ([]report.SearchResultOrder, error)
	SearchMenuItemsByKeywords(ctx context.Context, keywords []string, minPrice, maxPrice *float64) ([]report.SearchResultMenuItem, error)
	SearchOrdersByKeywords(ctx context.Context, keywords []string, minPrice, maxPrice *float64) ([]report.SearchResultOrder, error)
}
