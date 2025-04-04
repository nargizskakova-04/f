package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"frappuccino/internal/entity"

	"github.com/lib/pq"
)

type MenuRepository struct {
	db *sql.DB
}

func NewMenuRepository(db *sql.DB) *MenuRepository {
	return &MenuRepository{
		db: db,
	}
}

// TO-DO: для чего нужен контекст, почему тут int64
func (repo *MenuRepository) CreateMenuItem(ctx context.Context, menuItem entity.MenuItem) (string, error) {
	var ID string
	query := `
	INSERT INTO menu_items (
		name, 
		description, 
		price, 
		categories, 
		allergens, 
		size, 
		customization_options
	) VALUES ($1, $2, $3, $4, $5, $6, $7) 
	RETURNING menu_item_id;
	`

	// Convert Go slices to PostgreSQL arrays using pq.Array
	err := repo.db.QueryRowContext(ctx, query,
		menuItem.Name,
		menuItem.Description,
		menuItem.Price,
		pq.Array(menuItem.Categories), // Use pq.Array for []string
		pq.Array(menuItem.Allergens),  // Use pq.Array for []string
		menuItem.Size,
		menuItem.CustomizationOptions).Scan(&ID)

	return ID, err
}

func (repo *MenuRepository) GetMenuItem(ctx context.Context) ([]entity.MenuItem, error) {
	var menuItems []entity.MenuItem
	query := `
	SELECT 
		menu_item_id, 
		name, 
		description, 
		price, 
		categories, 
		allergens, 
		size, 
		customization_options, 
		updated_at
	FROM menu_items
	ORDER BY name
	`

	rows, err := repo.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var menu entity.MenuItem

		// Use temporary variables for categories and allergens
		var categories, allergens []string

		if err := rows.Scan(
			&menu.MenuItemID,
			&menu.Name,
			&menu.Description,
			&menu.Price,
			pq.Array(&categories), // Use pq.Array for scanning array types
			pq.Array(&allergens),  // Use pq.Array for scanning array types
			&menu.Size,
			&menu.CustomizationOptions,
			&menu.UpdatedAt,
		); err != nil {
			return nil, err
		}

		// Assign the scanned arrays to the menu struct
		menu.Categories = categories
		menu.Allergens = allergens

		menuItems = append(menuItems, menu)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return menuItems, nil
}

func (repo *MenuRepository) GetMenuByID(ctx context.Context, id string) (entity.MenuItem, error) {
	var menu entity.MenuItem
	query := `
    	SELECT 
		menu_item_id, 
		name, 
		description, 
		price, 
		categories, 
		allergens, 
		size, 
		customization_options, 
		updated_at
	FROM menu_items
    WHERE menu_item_id = $1;
    `
	var categories, allergens []string

	err := repo.db.QueryRowContext(ctx, query, id).Scan(
		&menu.MenuItemID,
		&menu.Name,
		&menu.Description,
		&menu.Price,
		pq.Array(&categories), // Use pq.Array for scanning array types
		pq.Array(&allergens),  // Use pq.Array for scanning array types
		&menu.Size,
		&menu.CustomizationOptions,
		&menu.UpdatedAt,
	)

	return menu, err
}

func (repo *MenuRepository) DeleteMenu(ctx context.Context, id string) (string, error) {
	query := `
 	DELETE
 	FROM menu_items
	WHERE menu_item_id = $1;
 	`
	_, err := repo.db.QueryContext(ctx, query, id)

	return id, err
}

func (r *MenuRepository) UpdateMenu(ctx context.Context, updates map[string]interface{}, id string) (string, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var oldPrice float64
	err = tx.QueryRowContext(ctx, "SELECT price FROM menu_items WHERE menu_item_id = $1", id).Scan(&oldPrice)
	if err != nil {
		return "", fmt.Errorf("get old price: %w", err)
	}

	if newVal, ok := updates["price"]; ok {
		var newPrice float64
		switch v := newVal.(type) {
		case float64:
			newPrice = v
		case float32:
			newPrice = float64(v)
		case int:
			newPrice = float64(v)
		default:
			return "", fmt.Errorf("invalid price type: %T", newVal)
		}
		if newPrice != oldPrice {
			_, err := tx.ExecContext(ctx, `
                INSERT INTO price_history (menu_item_id, old_price, new_price, changed_at, change_reason)
                VALUES ($1, $2, $3, CURRENT_TIMESTAMP, $4)
            `, id, oldPrice, newPrice, "manual update")
			if err != nil {
				return "", fmt.Errorf("insert price history: %w", err)
			}
		}
	}

	queryBuilder := strings.Builder{}
	queryBuilder.WriteString("UPDATE menu_items SET ")

	var args []interface{}
	i := 1
	for field, val := range updates {
		if i > 1 {
			queryBuilder.WriteString(", ")
		}
		queryBuilder.WriteString(fmt.Sprintf("%s = $%d", field, i))
		args = append(args, val)
		i++
	}
	queryBuilder.WriteString(fmt.Sprintf(" WHERE menu_item_id = $%d", i))
	args = append(args, id)

	if _, err := tx.ExecContext(ctx, queryBuilder.String(), args...); err != nil {
		return "", fmt.Errorf("update menu item: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit: %w", err)
	}

	return id, nil
}

func (r *MenuRepository) CreateMenuItemIngredients(ctx context.Context, menuItemID string, ingredients []entity.MenuItemIngredient) error {
	for _, ing := range ingredients {
		query := `
			INSERT INTO menu_item_ingredients (menu_item_id, ingredient_id, quantity, unit)
			VALUES ($1, $2, $3, $4)
		`
		_, err := r.db.ExecContext(ctx, query, menuItemID, ing.IngredientID, ing.Quantity, ing.Unit)
		if err != nil {
			return fmt.Errorf("insert ingredient: %w", err)
		}
	}
	return nil
}
func (repo *MenuRepository) GetAllPriceHistory(ctx context.Context) ([]entity.PriceHistory, error) {
	var priceHistory []entity.PriceHistory
	query := `
    SELECT 
        id,
        menu_item_id,
        old_price,
        new_price,
        changed_at,
        change_reason
    FROM price_history
    ORDER BY changed_at DESC
    `

	rows, err := repo.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var history entity.PriceHistory

		if err := rows.Scan(
			&history.ID,
			&history.MenuItemID,
			&history.OldPrice,
			&history.NewPrice,
			&history.ChangedAt,
			&history.ChangeReason,
		); err != nil {
			return nil, err
		}

		priceHistory = append(priceHistory, history)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return priceHistory, nil
}
func (repo *MenuRepository) GetMenuItemIngredients(ctx context.Context, menuItemID string) ([]entity.MenuItemIngredient, error) {
	query := `
		SELECT 
			menu_item_ingr_id, 
			menu_item_id, 
			ingredient_id, 
			quantity, 
			unit
		FROM menu_item_ingredients
		WHERE menu_item_id = $1
	`

	rows, err := repo.db.QueryContext(ctx, query, menuItemID)
	if err != nil {
		return nil, fmt.Errorf("failed to query menu item ingredients: %w", err)
	}
	defer rows.Close()

	var ingredients []entity.MenuItemIngredient
	for rows.Next() {
		var ingredient entity.MenuItemIngredient
		err := rows.Scan(
			&ingredient.ID,
			&ingredient.MenuItemID,
			&ingredient.IngredientID,
			&ingredient.Quantity,
			&ingredient.Unit,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan menu item ingredient: %w", err)
		}
		ingredients = append(ingredients, ingredient)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return ingredients, nil
}
