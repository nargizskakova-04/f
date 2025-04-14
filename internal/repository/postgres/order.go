package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"frappuccino/internal/dto/report"
	"frappuccino/internal/entity"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{
		db: db,
	}
}

func (repo *OrderRepository) CreateOrder(ctx context.Context, order entity.Order, items []entity.OrderItem) (string, error) {
	var orderID string
	orderQuery := `
		INSERT INTO orders (customer_name, special_instructions, total_amount, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING order_id;
	`
	err := repo.db.QueryRowContext(ctx, orderQuery,
		order.CustomerName,
		order.SpecialInstructions,
		order.TotalAmount,
		order.Status,
		order.CreatedAt,
		order.UpdatedAt,
	).Scan(&orderID)
	if err != nil {
		return "", fmt.Errorf("insert order: %w", err)
	}

	itemQuery := `
		INSERT INTO order_items (order_id, menu_item_id, quantity, price_at_time, customizations)
		VALUES ($1, $2, $3, $4, $5)
	`
	for _, item := range items {
		_, err := repo.db.ExecContext(ctx, itemQuery,
			orderID,
			item.MenuItemID,
			item.Quantity,
			item.PriceAtTime,
			item.Customizations,
		)
		if err != nil {
			return "", fmt.Errorf("insert order item: %w", err)
		}
	}

	return orderID, nil
}

func (repo *OrderRepository) GetMenuItemPrice(ctx context.Context, menuItemID string) (float64, error) {
	var price float64
	query := `SELECT price FROM menu_items WHERE menu_item_id = $1`

	err := repo.db.QueryRowContext(ctx, query, menuItemID).Scan(&price)
	if err != nil {
		return 0, fmt.Errorf("get price: %w", err)
	}

	return price, nil
}

func (repo *OrderRepository) GetOrderByID(ctx context.Context, orderID string) (entity.Order, error) {
	var o entity.Order
	query := `
	SELECT order_id, customer_name, special_instructions, total_amount, status, created_at, updated_at
	FROM orders
	WHERE order_id = $1;
	`
	err := repo.db.QueryRowContext(ctx, query, orderID).Scan(
		&o.OrderID,
		&o.CustomerName,
		&o.SpecialInstructions,
		&o.TotalAmount,
		&o.Status,
		&o.CreatedAt,
		&o.UpdatedAt,
	)
	return o, err
}

func (repo *OrderRepository) GetOrderItemsByOrderID(ctx context.Context, orderID string) ([]entity.OrderItem, error) {
	query := `
	SELECT menu_item_id, quantity, price_at_time, customizations
	FROM order_items
	WHERE order_id = $1;
	`

	rows, err := repo.db.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []entity.OrderItem
	var customizationsNullable sql.NullString
	for rows.Next() {
		var i entity.OrderItem
		if err := rows.Scan(&i.MenuItemID, &i.Quantity, &i.PriceAtTime, &customizationsNullable); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

func (repo *OrderRepository) GetAllOrders(ctx context.Context) ([]entity.Order, error) {
	var orders []entity.Order
	query := `
        SELECT
            order_id,
            customer_name,
            special_instructions,
            total_amount,
            status,
            created_at,
            updated_at
        FROM orders
        ORDER BY created_at DESC
    `

	rows, err := repo.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var specialInstructionsNullable sql.NullString

	for rows.Next() {
		var order entity.Order
		if err := rows.Scan(
			&order.OrderID,
			&order.CustomerName,
			&specialInstructionsNullable,
			&order.TotalAmount,
			&order.Status,
			&order.CreatedAt,
			&order.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if specialInstructionsNullable.Valid {
			order.SpecialInstructions = json.RawMessage(specialInstructionsNullable.String)
		} else {
			order.SpecialInstructions = json.RawMessage(`{}`)
		}

		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (repo *OrderRepository) UpdateOrder(ctx context.Context, orderID string, updates map[string]interface{}) error {
	var changeReason string
	if reasonVal, hasReason := updates["change_reason"]; hasReason {
		changeReason = reasonVal.(string)
		delete(updates, "change_reason")
	}

	statusUpdate, hasStatusUpdate := updates["status"]
	if hasStatusUpdate {
		var oldStatus string
		getStatusQuery := `SELECT status FROM orders WHERE order_id = $1`
		err := repo.db.QueryRowContext(ctx, getStatusQuery, orderID).Scan(&oldStatus)
		if err != nil {
			return fmt.Errorf("get current status: %w", err)
		}

		historyQuery := `
            INSERT INTO order_status_history 
            (order_id, old_status, new_status, change_reason)
            VALUES ($1, $2, $3, $4)
        `
		_, err = repo.db.ExecContext(ctx, historyQuery,
			orderID, oldStatus, statusUpdate.(string), changeReason)
		if err != nil {
			return fmt.Errorf("insert status history: %w", err)
		}
	}

	if len(updates) == 0 {
		return nil
	}

	queryBuilder := strings.Builder{}
	queryBuilder.WriteString("UPDATE orders SET ")

	updates["updated_at"] = time.Now()

	values := []interface{}{}

	paramIndex := 1

	isFirst := true

	for field, value := range updates {
		if !isFirst {
			queryBuilder.WriteString(", ")
		}
		queryBuilder.WriteString(field + " = $" + strconv.Itoa(paramIndex))
		values = append(values, value)
		paramIndex++
		isFirst = false
	}

	queryBuilder.WriteString(" WHERE order_id = $" + strconv.Itoa(paramIndex))
	values = append(values, orderID)

	_, err := repo.db.ExecContext(ctx, queryBuilder.String(), values...)
	if err != nil {
		return fmt.Errorf("update order: %w", err)
	}

	return nil
}

func (repo *OrderRepository) GetAllOrderStatusHistory(ctx context.Context) ([]entity.OrderStatusHistory, error) {
	query := `
        SELECT 
            order_status_id, 
            order_id, 
            old_status, 
            new_status, 
            changed_at, 
            change_reason
        FROM order_status_history
        ORDER BY changed_at DESC
    `

	rows, err := repo.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query all status history: %w", err)
	}
	defer rows.Close()

	var history []entity.OrderStatusHistory
	for rows.Next() {
		var h entity.OrderStatusHistory
		if err := rows.Scan(
			&h.OrderStatusID,
			&h.OrderID,
			&h.OldStatus,
			&h.NewStatus,
			&h.ChangedAt,
			&h.ChangeReason,
		); err != nil {
			return nil, fmt.Errorf("scan status history: %w", err)
		}
		history = append(history, h)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate status history: %w", err)
	}

	return history, nil
}

func (repo *OrderRepository) DeleteOrder(ctx context.Context, id string) (string, error) {
	query := `
 	DELETE
 	FROM orders
	WHERE order_id = $1;
 	`
	_, err := repo.db.QueryContext(ctx, query, id)

	return id, err
}

func (repo *OrderRepository) GetNumberOfOrderedItems(
	ctx context.Context,
	startDate, endDate *time.Time,
) (map[string]int, error) {
	query := `
		SELECT m.name, SUM(oi.quantity) as total_quantity
		FROM order_items oi
		JOIN orders o ON oi.order_id = o.order_id
		JOIN menu_items m ON oi.menu_item_id = m.menu_item_id
		WHERE 1=1
	`

	var args []interface{}
	var argIndex int = 1

	if startDate != nil {
		query += ` AND o.created_at >= $` + repo.nextArgIndex(&argIndex)
		args = append(args, startDate)
	}

	if endDate != nil {
		endDatePlusDay := endDate.AddDate(0, 0, 1)
		query += ` AND o.created_at < $` + repo.nextArgIndex(&argIndex)
		args = append(args, endDatePlusDay)
	}

	query += `
		GROUP BY m.name
		ORDER BY m.name
	`

	rows, err := repo.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int)

	for rows.Next() {
		var name string
		var quantity int

		if err := rows.Scan(&name, &quantity); err != nil {
			return nil, err
		}

		result[name] = quantity
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (repo *OrderRepository) nextArgIndex(index *int) string {
	current := *index
	*index++
	return string(rune('0' + current))
}

func (repo *OrderRepository) GetOrderedItemsByDay(ctx context.Context, month time.Month, year int) ([]report.DayCount, error) {
	startDate := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)

	query := `
		SELECT 
			EXTRACT(DAY FROM o.created_at) as day,
			SUM(oi.quantity) as count
		FROM 
			orders o
		JOIN 
			order_items oi ON o.order_id = oi.order_id
		WHERE 
			o.created_at >= $1 AND o.created_at < $2
			AND o.status != 'cancelled'
		GROUP BY 
			EXTRACT(DAY FROM o.created_at)
		ORDER BY 
			day
	`

	rows, err := repo.db.QueryContext(ctx, query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []report.DayCount
	for rows.Next() {
		var dc report.DayCount
		if err := rows.Scan(&dc.Day, &dc.Count); err != nil {
			return nil, err
		}
		results = append(results, dc)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (repo *OrderRepository) GetOrderedItemsByMonth(ctx context.Context, year int) ([]report.MonthCount, error) {
	startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(1, 0, 0)

	query := `
		SELECT 
			EXTRACT(MONTH FROM o.created_at) as month,
			SUM(oi.quantity) as count
		FROM 
			orders o
		JOIN 
			order_items oi ON o.order_id = oi.order_id
		WHERE 
			o.created_at >= $1 AND o.created_at < $2
			AND o.status != 'cancelled'
		GROUP BY 
			EXTRACT(MONTH FROM o.created_at)
		ORDER BY 
			month
	`

	rows, err := repo.db.QueryContext(ctx, query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []report.MonthCount
	monthNames := map[int]string{
		1:  "january",
		2:  "february",
		3:  "march",
		4:  "april",
		5:  "may",
		6:  "june",
		7:  "july",
		8:  "august",
		9:  "september",
		10: "october",
		11: "november",
		12: "december",
	}

	for rows.Next() {
		var monthNum int
		var count int
		if err := rows.Scan(&monthNum, &count); err != nil {
			return nil, err
		}

		results = append(results, report.MonthCount{
			Month: monthNames[monthNum],
			Count: count,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}
