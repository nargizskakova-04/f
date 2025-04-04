package report

import (
	"context"
	"time"

	"frappuccino/internal/dto/report"
)

type searchRepo interface {
	SearchMenuItems(ctx context.Context, query string, minPrice, maxPrice *float64) ([]report.SearchResultMenuItem, error)
	SearchOrders(ctx context.Context, query string, minPrice, maxPrice *float64) ([]report.SearchResultOrder, error)
	SearchMenuItemsByKeywords(ctx context.Context, keywords []string, minPrice, maxPrice *float64) ([]report.SearchResultMenuItem, error)
	SearchOrdersByKeywords(ctx context.Context, keywords []string, minPrice, maxPrice *float64) ([]report.SearchResultOrder, error)
}

type orderRepo interface {
	GetOrderedItemsByDay(ctx context.Context, month time.Month, year int) ([]report.DayCount, error)
	GetOrderedItemsByMonth(ctx context.Context, year int) ([]report.MonthCount, error)
}
