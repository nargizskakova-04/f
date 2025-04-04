package postgres

import (
	"context"
	"fmt"
	"time"

	"frappuccino/internal/dto/report"
)

// GetTotalSales returns the total sales amount for the given date range and status
func (repo *OrderRepository) GetTotalSales(ctx context.Context, startDate, endDate *time.Time, status string) (float64, int, error) {
	query := `
		SELECT 
			COALESCE(SUM(total_amount), 0) as total_sales,
			COUNT(*) as order_count
		FROM orders
		WHERE 1=1
	`

	var args []interface{}
	var argIndex int = 1

	// Add date filters if provided
	if startDate != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, startDate)
		argIndex++
	}

	if endDate != nil {
		// Add one day to include the end date in the results (until end of the day)
		endDatePlusDay := endDate.AddDate(0, 0, 1)
		query += fmt.Sprintf(" AND created_at < $%d", argIndex)
		args = append(args, endDatePlusDay)
		argIndex++
	}

	// Add status filter if provided
	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, status)
	}

	// Execute the query
	var totalSales float64
	var orderCount int
	err := repo.db.QueryRowContext(ctx, query, args...).Scan(&totalSales, &orderCount)
	if err != nil {
		return 0, 0, fmt.Errorf("error querying total sales: %w", err)
	}

	return totalSales, orderCount, nil
}

// GetPopularItems returns the most popular menu items for the given date range
func (repo *OrderRepository) GetPopularItems(ctx context.Context, startDate, endDate *time.Time, limit int) ([]report.PopularItem, int, float64, error) {
	// Default limit if not specified
	if limit <= 0 {
		limit = 10
	}

	query := `
		WITH item_sales AS (
			SELECT 
				m.menu_item_id,
				m.name,
				SUM(oi.quantity) as quantity_sold,
				SUM(oi.quantity * oi.price_at_time) as total_revenue
			FROM 
				order_items oi
			JOIN 
				orders o ON oi.order_id = o.order_id
			JOIN 
				menu_items m ON oi.menu_item_id = m.menu_item_id
			WHERE 1=1
	`

	var args []interface{}
	var argIndex int = 1

	// Add date filters if provided
	if startDate != nil {
		query += fmt.Sprintf(" AND o.created_at >= $%d", argIndex)
		args = append(args, startDate)
		argIndex++
	}

	if endDate != nil {
		// Add one day to include the end date in the results (until end of the day)
		endDatePlusDay := endDate.AddDate(0, 0, 1)
		query += fmt.Sprintf(" AND o.created_at < $%d", argIndex)
		args = append(args, endDatePlusDay)
		argIndex++
	}

	// Exclude cancelled orders
	query += " AND o.status != 'cancelled'"

	// Group by menu item, sort by quantity sold in descending order, and limit the results
	query += `
			GROUP BY 
				m.menu_item_id, m.name
		)
		SELECT 
			menu_item_id,
			name,
			quantity_sold,
			total_revenue
		FROM 
			item_sales
		ORDER BY 
			quantity_sold DESC, total_revenue DESC
		LIMIT $` + fmt.Sprintf("%d", argIndex)

	args = append(args, limit)

	// Execute the query
	rows, err := repo.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("error querying popular items: %w", err)
	}
	defer rows.Close()

	// Process the results
	var items []report.PopularItem
	var totalQuantity int
	var totalRevenue float64

	for rows.Next() {
		var item report.PopularItem
		if err := rows.Scan(&item.MenuItemID, &item.Name, &item.Quantity, &item.TotalRevenue); err != nil {
			return nil, 0, 0, fmt.Errorf("error scanning popular item row: %w", err)
		}
		items = append(items, item)

		// Accumulate totals
		totalQuantity += item.Quantity
		totalRevenue += item.TotalRevenue
	}

	if err := rows.Err(); err != nil {
		return nil, 0, 0, fmt.Errorf("error iterating popular item rows: %w", err)
	}

	// Calculate percentages if there are any items
	if totalRevenue > 0 {
		for i := range items {
			items[i].PercentOfSales = (items[i].TotalRevenue / totalRevenue) * 100
		}
	}

	return items, totalQuantity, totalRevenue, nil
}
