package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"frappuccino/internal/dto/report"
)

type SearchRepository struct {
	db *sql.DB
}

func NewSearchRepository(db *sql.DB) *SearchRepository {
	return &SearchRepository{
		db: db,
	}
}

// SearchMenuItems searches menu items based on search criteria
func (repo *SearchRepository) SearchMenuItems(
	ctx context.Context,
	query string,
	minPrice, maxPrice *float64,
) ([]report.SearchResultMenuItem, error) {
	// Build the query
	sqlQuery := `
        SELECT 
            menu_item_id,
            name,
            description,
            price,
            ts_rank(to_tsvector('english', name || ' ' || COALESCE(description, '')), plainto_tsquery('english', $1)) AS relevance
        FROM menu_items
        WHERE to_tsvector('english', name || ' ' || COALESCE(description, '')) @@ plainto_tsquery('english', $1)
    `

	// Add price filters if provided
	args := []interface{}{query}
	paramIndex := 2 // starting from second parameter since $1 is already used

	if minPrice != nil {
		sqlQuery += fmt.Sprintf(" AND price >= $%d", paramIndex)
		args = append(args, *minPrice)
		paramIndex++
	}

	if maxPrice != nil {
		sqlQuery += fmt.Sprintf(" AND price <= $%d", paramIndex)
		args = append(args, *maxPrice)
	}

	// Order by relevance and limit results
	sqlQuery += " ORDER BY relevance DESC LIMIT 20"

	// Execute query
	rows, err := repo.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("error searching menu items: %w", err)
	}
	defer rows.Close()

	var results []report.SearchResultMenuItem
	for rows.Next() {
		var item report.SearchResultMenuItem
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Description,
			&item.Price,
			&item.Relevance,
		); err != nil {
			return nil, fmt.Errorf("error scanning menu item search result: %w", err)
		}
		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating menu item search results: %w", err)
	}

	return results, nil
}

// SearchOrders searches orders based on search criteria
func (repo *SearchRepository) SearchOrders(
	ctx context.Context,
	query string,
	minPrice, maxPrice *float64,
) ([]report.SearchResultOrder, error) {
	// Build the query
	sqlQuery := `
        WITH matching_orders AS (
            SELECT 
                o.order_id,
                o.customer_name,
                o.total_amount,
                ts_rank(to_tsvector('english', o.customer_name || ' ' || COALESCE(o.special_instructions::text, '')), plainto_tsquery('english', $1)) AS relevance
            FROM orders o
            WHERE to_tsvector('english', o.customer_name || ' ' || COALESCE(o.special_instructions::text, '')) @@ plainto_tsquery('english', $1)
        )
        SELECT 
            mo.order_id,
            mo.customer_name,
            mo.total_amount,
            mo.relevance,
            array_agg(mi.name) as item_names
        FROM matching_orders mo
        JOIN order_items oi ON mo.order_id = oi.order_id
        JOIN menu_items mi ON oi.menu_item_id = mi.menu_item_id
    `

	// Add price filters if provided
	args := []interface{}{query}
	paramIndex := 2 // starting from second parameter since $1 is already used

	if minPrice != nil {
		sqlQuery += fmt.Sprintf(" WHERE mo.total_amount >= $%d", paramIndex)
		args = append(args, *minPrice)
		paramIndex++
	} else {
		sqlQuery += " WHERE 1=1"
	}

	if maxPrice != nil {
		sqlQuery += fmt.Sprintf(" AND mo.total_amount <= $%d", paramIndex)
		args = append(args, *maxPrice)
	}

	// Group by order details and order by relevance
	sqlQuery += `
        GROUP BY mo.order_id, mo.customer_name, mo.total_amount, mo.relevance
        ORDER BY mo.relevance DESC
        LIMIT 20
    `

	// Execute query
	rows, err := repo.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("error searching orders: %w", err)
	}
	defer rows.Close()

	var results []report.SearchResultOrder
	for rows.Next() {
		var order report.SearchResultOrder
		var itemsArray []string

		if err := rows.Scan(
			&order.ID,
			&order.CustomerName,
			&order.Total,
			&order.Relevance,
			&itemsArray,
		); err != nil {
			return nil, fmt.Errorf("error scanning order search result: %w", err)
		}

		order.Items = itemsArray
		results = append(results, order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating order search results: %w", err)
	}

	return results, nil
}

// SearchMenuItemsByKeywords searches menu items using individual keywords
func (repo *SearchRepository) SearchMenuItemsByKeywords(
	ctx context.Context,
	keywords []string,
	minPrice, maxPrice *float64,
) ([]report.SearchResultMenuItem, error) {
	// Transform keywords into a tsquery format (word1 | word2 | word3)
	tsquery := strings.Join(keywords, " | ")

	// Build the query
	sqlQuery := `
        SELECT 
            menu_item_id,
            name,
            description,
            price,
            ts_rank(to_tsvector('english', name || ' ' || COALESCE(description, '')), to_tsquery('english', $1)) AS relevance
        FROM menu_items
        WHERE to_tsvector('english', name || ' ' || COALESCE(description, '')) @@ to_tsquery('english', $1)
    `

	// Add price filters if provided
	args := []interface{}{tsquery}
	paramIndex := 2 // starting from second parameter since $1 is already used

	if minPrice != nil {
		sqlQuery += fmt.Sprintf(" AND price >= $%d", paramIndex)
		args = append(args, *minPrice)
		paramIndex++
	}

	if maxPrice != nil {
		sqlQuery += fmt.Sprintf(" AND price <= $%d", paramIndex)
		args = append(args, *maxPrice)
	}

	// Order by relevance and limit results
	sqlQuery += " ORDER BY relevance DESC LIMIT 20"

	// Execute query
	rows, err := repo.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("error searching menu items by keywords: %w", err)
	}
	defer rows.Close()

	var results []report.SearchResultMenuItem
	for rows.Next() {
		var item report.SearchResultMenuItem
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Description,
			&item.Price,
			&item.Relevance,
		); err != nil {
			return nil, fmt.Errorf("error scanning menu item search result: %w", err)
		}
		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating menu item search results: %w", err)
	}

	return results, nil
}

// SearchOrdersByKeywords searches orders using individual keywords
func (repo *SearchRepository) SearchOrdersByKeywords(
	ctx context.Context,
	keywords []string,
	minPrice, maxPrice *float64,
) ([]report.SearchResultOrder, error) {
	// Transform keywords into a tsquery format (word1 | word2 | word3)
	tsquery := strings.Join(keywords, " | ")

	// Build the query
	sqlQuery := `
        WITH matching_orders AS (
            SELECT 
                o.order_id,
                o.customer_name,
                o.total_amount,
                ts_rank(to_tsvector('english', o.customer_name || ' ' || COALESCE(o.special_instructions::text, '')), to_tsquery('english', $1)) AS relevance
            FROM orders o
            WHERE to_tsvector('english', o.customer_name || ' ' || COALESCE(o.special_instructions::text, '')) @@ to_tsquery('english', $1)
        )
        SELECT 
            mo.order_id,
            mo.customer_name,
            mo.total_amount,
            mo.relevance,
            array_agg(mi.name) as item_names
        FROM matching_orders mo
        JOIN order_items oi ON mo.order_id = oi.order_id
        JOIN menu_items mi ON oi.menu_item_id = mi.menu_item_id
    `

	// Add price filters if provided
	args := []interface{}{tsquery}
	paramIndex := 2 // starting from second parameter since $1 is already used

	if minPrice != nil {
		sqlQuery += fmt.Sprintf(" WHERE mo.total_amount >= $%d", paramIndex)
		args = append(args, *minPrice)
		paramIndex++
	} else {
		sqlQuery += " WHERE 1=1"
	}

	if maxPrice != nil {
		sqlQuery += fmt.Sprintf(" AND mo.total_amount <= $%d", paramIndex)
		args = append(args, *maxPrice)
	}

	// Group by order details and order by relevance
	sqlQuery += `
        GROUP BY mo.order_id, mo.customer_name, mo.total_amount, mo.relevance
        ORDER BY mo.relevance DESC
        LIMIT 20
    `

	// Execute query
	rows, err := repo.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("error searching orders by keywords: %w", err)
	}
	defer rows.Close()

	var results []report.SearchResultOrder
	for rows.Next() {
		var order report.SearchResultOrder
		var itemsArray []string

		if err := rows.Scan(
			&order.ID,
			&order.CustomerName,
			&order.Total,
			&order.Relevance,
			&itemsArray,
		); err != nil {
			return nil, fmt.Errorf("error scanning order search result: %w", err)
		}

		order.Items = itemsArray
		results = append(results, order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating order search results: %w", err)
	}

	return results, nil
}
