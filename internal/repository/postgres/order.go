package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

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

// CreateOrder inserts order and order items inside a transaction.
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

// GetMenuItemPrice gets current price of the menu item
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
			order.SpecialInstructions = json.RawMessage(`{}`) // Пустой JSON объект вместо NULL
		}

		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (repo *OrderRepository) UpdateOrder(ctx context.Context, orderID string, updates map[string]interface{}) error {
	// Извлекаем и сохраняем change_reason, если есть
	var changeReason string
	if reasonVal, hasReason := updates["change_reason"]; hasReason {
		changeReason = reasonVal.(string)
		// Удаляем change_reason из updates, так как это не столбец в таблице orders
		delete(updates, "change_reason")
	}

	// Проверяем, обновляется ли статус
	statusUpdate, hasStatusUpdate := updates["status"]
	if hasStatusUpdate {
		// Если статус обновляется, нам нужно обработать его отдельно
		// для записи в order_status_history
		var oldStatus string
		getStatusQuery := `SELECT status FROM orders WHERE order_id = $1`
		err := repo.db.QueryRowContext(ctx, getStatusQuery, orderID).Scan(&oldStatus)
		if err != nil {
			return fmt.Errorf("get current status: %w", err)
		}

		// Добавляем запись в order_status_history
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

	// Продолжаем только если есть поля для обновления
	// после потенциального удаления change_reason
	if len(updates) == 0 {
		return nil
	}

	// Начинаем строить запрос
	queryBuilder := strings.Builder{}
	queryBuilder.WriteString("UPDATE orders SET ")

	// Всегда обновляем поле updated_at
	updates["updated_at"] = time.Now()

	// Значения для передачи в запрос
	values := []interface{}{}

	// Отслеживаем индекс параметра
	paramIndex := 1

	// Отслеживаем, является ли это первым полем (для расстановки запятых)
	isFirst := true

	// Добавляем поля, которые нужно обновить
	for field, value := range updates {
		if !isFirst {
			queryBuilder.WriteString(", ")
		}
		queryBuilder.WriteString(field + " = $" + strconv.Itoa(paramIndex))
		values = append(values, value)
		paramIndex++
		isFirst = false
	}

	// Добавляем условие WHERE и параметр id
	queryBuilder.WriteString(" WHERE order_id = $" + strconv.Itoa(paramIndex))
	values = append(values, orderID)

	// Выполняем запрос
	_, err := repo.db.ExecContext(ctx, queryBuilder.String(), values...)
	if err != nil {
		return fmt.Errorf("update order: %w", err)
	}

	return nil
}

// Add to OrderRepository
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
