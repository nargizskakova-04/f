package postgres

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	"frappuccino/internal/entity"
)

type InventoryRepository struct {
	db *sql.DB
}

func NewInventoryRepository(db *sql.DB) *InventoryRepository {
	return &InventoryRepository{
		db: db,
	}
}

// TO-DO: для чего нужен контекст, почему тут int64
func (repo *InventoryRepository) CreateInventory(ctx context.Context, inventory entity.Inventory) (string, error) {
	var ID string
	query := `
	 INSERT INTO inventory (name, quantity, unit, unit_price, reorder_point) VALUES ($1, $2, $3, $4, $5) RETURNING ingredient_id;
	  `
	err := repo.db.QueryRowContext(ctx, query,
		inventory.Name,
		inventory.Quantity,
		inventory.Unit,
		inventory.UnitPrice,
		inventory.ReorderPoint).Scan(&ID)
	return ID, err
}

func (repo *InventoryRepository) GetInventory(ctx context.Context) ([]entity.Inventory, error) {
	var inventories []entity.Inventory
	query := `
	SELECT * 
	FROM inventory
	ORDER BY name
	`

	rows, err := repo.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var inv entity.Inventory
		if err := rows.Scan(
			&inv.IngredientID,
			&inv.Name,
			&inv.Quantity,
			&inv.Unit,
			&inv.UnitPrice,
			&inv.ReorderPoint,
			&inv.LastUpdated,
		); err != nil {
			return nil, err
		}
		inventories = append(inventories, inv)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return inventories, nil
}

func (repo *InventoryRepository) GetInventoryByID(ctx context.Context, id string) (entity.Inventory, error) {
	var inv entity.Inventory
	query := `
    SELECT ingredient_id, name, quantity, unit, unit_price, reorder_point, last_updated 
    FROM inventory 
    WHERE ingredient_id = $1;
    `

	err := repo.db.QueryRowContext(ctx, query, id).Scan(
		&inv.IngredientID,
		&inv.Name,
		&inv.Quantity,
		&inv.Unit,
		&inv.UnitPrice,
		&inv.ReorderPoint,
		&inv.LastUpdated,
	)

	return inv, err
}

func (repo *InventoryRepository) DeleteInventory(ctx context.Context, id string) (string, error) {
	query := `
	DELETE
	FROM inventory
	WHERE ingredient_id = $1;
	`
	_, err := repo.db.QueryContext(ctx, query, id)

	return id, err
}

func (repo *InventoryRepository) UpdateInventory(ctx context.Context, updates map[string]interface{}, id string) (string, error) {
	// Start building the query
	queryBuilder := strings.Builder{}
	queryBuilder.WriteString("UPDATE inventory SET ")

	// Values to be passed to the query
	values := []interface{}{}

	// Keep track of the parameter index
	paramIndex := 1

	// Track if it's the first field (for comma placement)
	isFirst := true

	// Add fields that need to be updated
	for field, value := range updates {
		if !isFirst {
			queryBuilder.WriteString(", ")
		}

		queryBuilder.WriteString(field + " = $" + strconv.Itoa(paramIndex))
		values = append(values, value)
		paramIndex++
		isFirst = false
	}

	// Add the WHERE clause and the id parameter
	queryBuilder.WriteString(" WHERE ingredient_id = $" + strconv.Itoa(paramIndex))
	values = append(values, id)

	// Execute the query
	_, err := repo.db.ExecContext(ctx, queryBuilder.String(), values...)
	return id, err
}

func (repo *InventoryRepository) CreateInventoryTransaction(ctx context.Context, transaction entity.InventoryTransaction) error {
	query := `
        INSERT INTO inventory_transactions 
        (ingredient_id, quantity_change, transaction_type, reason) 
        VALUES ($1, $2, $3, $4);
    `
	_, err := repo.db.ExecContext(ctx, query,
		transaction.IngredientID,
		transaction.QuantityChange,
		transaction.TransactionType,
		transaction.Reason)

	return err
}

func (repo *InventoryRepository) GetInventoryTransactions(ctx context.Context, ingredientID string) ([]entity.InventoryTransaction, error) {
	var transactions []entity.InventoryTransaction

	query := `
        SELECT transaction_id, ingredient_id, quantity_change, transaction_type, reason, created_at
        FROM inventory_transactions
        WHERE ingredient_id = $1
        ORDER BY created_at DESC
    `

	rows, err := repo.db.QueryContext(ctx, query, ingredientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tx entity.InventoryTransaction
		if err := rows.Scan(
			&tx.TransactionID,
			&tx.IngredientID,
			&tx.QuantityChange,
			&tx.TransactionType,
			&tx.Reason,
			&tx.CreatedAt,
		); err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

func (repo *InventoryRepository) GetLeftOvers(ctx context.Context, sortBy string, page, pageSize int) ([]entity.Inventory, int, error) {
	// Build the query with sorting and pagination
	queryBuilder := strings.Builder{}
	queryBuilder.WriteString(`
		SELECT ingredient_id, name, quantity, unit, unit_price, reorder_point, last_updated
		FROM inventory
	`)

	// Add sorting
	switch sortBy {
	case "price":
		queryBuilder.WriteString(" ORDER BY unit_price DESC")
	case "quantity":
		queryBuilder.WriteString(" ORDER BY quantity DESC")
	default:
		queryBuilder.WriteString(" ORDER BY name")
	}

	// Add pagination
	offset := (page - 1) * pageSize
	queryBuilder.WriteString(" LIMIT $1 OFFSET $2")

	// Execute the query to get items
	rows, err := repo.db.QueryContext(ctx, queryBuilder.String(), pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	// Parse the inventory items
	var inventories []entity.Inventory
	for rows.Next() {
		var inv entity.Inventory
		if err := rows.Scan(
			&inv.IngredientID,
			&inv.Name,
			&inv.Quantity,
			&inv.Unit,
			&inv.UnitPrice,
			&inv.ReorderPoint,
			&inv.LastUpdated,
		); err != nil {
			return nil, 0, err
		}
		inventories = append(inventories, inv)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	// Get total count for pagination info
	var totalCount int
	countQuery := "SELECT COUNT(*) FROM inventory"
	err = repo.db.QueryRowContext(ctx, countQuery).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	return inventories, totalCount, nil
}
