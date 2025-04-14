package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"frappuccino/internal/entity"
)

type Transaction struct {
	tx *sql.Tx
}

func (repo *OrderRepository) Begin(ctx context.Context) (*Transaction, error) {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return &Transaction{tx: tx}, nil
}

func (t *Transaction) Commit() error {
	return t.tx.Commit()
}

func (t *Transaction) Rollback() error {
	return t.tx.Rollback()
}

func (repo *OrderRepository) CreateOrderWithTx(ctx context.Context, tx *Transaction, order entity.Order, items []entity.OrderItem) (string, error) {
	var orderID string
	orderQuery := `
		INSERT INTO orders (customer_name, special_instructions, total_amount, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING order_id;
	`
	err := tx.tx.QueryRowContext(ctx, orderQuery,
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
		_, err := tx.tx.ExecContext(ctx, itemQuery,
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

func (repo *InventoryRepository) UpdateInventoryWithTx(ctx context.Context, tx *Transaction, updates map[string]interface{}, id string) error {
	query, args := buildUpdateQuery("inventory", updates, "ingredient_id", id)
	_, err := tx.tx.ExecContext(ctx, query, args...)
	return err
}

func (repo *InventoryRepository) CreateInventoryTransactionWithTx(ctx context.Context, tx *Transaction, transaction entity.InventoryTransaction) error {
	query := `
        INSERT INTO inventory_transactions 
        (ingredient_id, quantity_change, transaction_type, reason) 
        VALUES ($1, $2, $3, $4);
    `
	_, err := tx.tx.ExecContext(ctx, query,
		transaction.IngredientID,
		transaction.QuantityChange,
		transaction.TransactionType,
		transaction.Reason)

	return err
}

func buildUpdateQuery(table string, updates map[string]interface{}, idField, idValue string) (string, []interface{}) {
	query := fmt.Sprintf("UPDATE %s SET ", table)
	var args []interface{}
	i := 1

	for field, value := range updates {
		if i > 1 {
			query += ", "
		}
		query += fmt.Sprintf("%s = $%d", field, i)
		args = append(args, value)
		i++
	}

	query += fmt.Sprintf(" WHERE %s = $%d", idField, i)
	args = append(args, idValue)

	return query, args
}
