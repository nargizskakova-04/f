package menu

import (
	"context"
	"errors"
	"frappuccino/internal/dto/menu"
	"frappuccino/internal/entity"
	"log"
	"time"
)

type MenuService struct {
	menuRepo menuRepo
	logger   *log.Logger
}

func NewMenuService(menuRepo menuRepo, logger *log.Logger) *MenuService {
	return &MenuService{
		menuRepo: menuRepo,
		logger:   logger,
	}
}

func (s *MenuService) CreateMenuItem(ctx context.Context, request menu.CreateMenuItemRequest) (string, error) {
	// Step 1: Build menu_item entity
	insertToDBMenuItem := entity.MenuItem{
		Name:                 request.Name,
		Description:          request.Description,
		Price:                request.Price,
		Categories:           request.Categories,
		Allergens:            request.Allergens,
		Size:                 request.Size,
		CustomizationOptions: request.CustomizationOptions,
		UpdatedAt:            time.Now(),
	}

	// Step 2: Insert menu_item
	id, err := s.menuRepo.CreateMenuItem(ctx, insertToDBMenuItem)
	if err != nil {
		s.logger.Println("CreateMenuItem error:", err)
		return "", err
	}

	// Step 3: Prepare ingredients for this menu_item
	var ingredients []entity.MenuItemIngredient
	for _, dtoIng := range request.Ingredients {
		ingredients = append(ingredients, entity.MenuItemIngredient{
			MenuItemID:   id,
			IngredientID: dtoIng.IngredientID,
			Quantity:     dtoIng.Quantity,
			Unit:         dtoIng.Unit,
		})
	}

	// Step 4: Insert menu_item_ingredients
	if err := s.menuRepo.CreateMenuItemIngredients(ctx, id, ingredients); err != nil {
		s.logger.Println("CreateMenuItemIngredients error:", err)
		return "", err
	}

	return id, nil
}

func (s *MenuService) GetMenuItem(ctx context.Context) ([]menu.GetMenuResponse, error) {
	// Call the repository function to get all menu items
	items, err := s.menuRepo.GetMenuItem(ctx)
	if err != nil {
		s.logger.Println("Error retrieving menu items:", err)
		return nil, err
	}

	// Map entity.Menu items to the response type
	var response []menu.GetMenuResponse
	for _, item := range items {
		response = append(response, menu.GetMenuResponse{
			MenuItemID:           item.MenuItemID,
			Name:                 item.Name,
			Description:          item.Description,
			Price:                item.Price,
			Categories:           item.Categories,
			Size:                 item.Size,
			CustomizationOptions: item.CustomizationOptions,
			UpdatedAt:            item.UpdatedAt,
		})
	}

	return response, nil
}

func (s *MenuService) GetMenuByID(ctx context.Context, id string) (menu.GetMenuResponse, error) {
	item, err := s.menuRepo.GetMenuByID(ctx, id)
	if err != nil {
		s.logger.Println("Error retrieving menu item:", err)
		return menu.GetMenuResponse{}, err
	}

	response := menu.GetMenuResponse{
		MenuItemID:           item.MenuItemID,
		Name:                 item.Name,
		Description:          item.Description,
		Price:                item.Price,
		Categories:           item.Categories,
		Size:                 item.Size,
		CustomizationOptions: item.CustomizationOptions,
		UpdatedAt:            item.UpdatedAt,
	}

	return response, nil
}

func (s *MenuService) DeleteMenu(ctx context.Context, id string) (string, error) {
	menu_item_id, err := s.menuRepo.DeleteMenu(ctx, id)
	if err != nil {
		s.logger.Println(err)
		return "", err
	}
	return menu_item_id, nil
}

func (s *MenuService) UpdateMenu(ctx context.Context, request menu.UpdateMenuRequest, id string) (string, error) {
	// Create a map to store only the fields that need updating
	updates := make(map[string]interface{})

	// Add fields to the update map only if they're provided
	if request.Name != nil {
		updates["name"] = *request.Name
	}

	if request.Description != nil {
		updates["description"] = *request.Description
	}

	if request.Price != nil {
		updates["price"] = *request.Price
	}

	if request.Categories != nil {
		updates["categories"] = *request.Categories
	}

	if request.Allergens != nil {
		updates["allergens"] = *request.Allergens
	}

	if request.Size != nil {
		updates["size"] = *request.Size
	}

	if request.CustomizationOptions != nil {
		updates["customization_options"] = *request.CustomizationOptions
	}
	// Always update the last_updated timestamp
	updates["updated_at"] = time.Now()

	// Don't proceed if there are no fields to update
	if len(updates) == 1 && updates["updated_at"] != nil {
		return "", errors.New("no fields to update")
	}

	// Call the repository with only the fields that need updating
	id, err := s.menuRepo.UpdateMenu(ctx, updates, id)
	if err != nil {
		s.logger.Println(err)
		return "", err
	}

	return id, nil
}

func (s *MenuService) GetAllPriceHistory(ctx context.Context) ([]menu.GetPriceHistoryResponse, error) {
	// Call the repository function to get all price history records
	histories, err := s.menuRepo.GetAllPriceHistory(ctx)
	if err != nil {
		s.logger.Println("Error retrieving price history:", err)
		return nil, err
	}

	// Map entity.PriceHistory items to the response type
	var response []menu.GetPriceHistoryResponse
	for _, history := range histories {
		response = append(response, menu.GetPriceHistoryResponse{
			ID:           history.ID,
			MenuItemID:   history.MenuItemID,
			OldPrice:     history.OldPrice,
			NewPrice:     history.NewPrice,
			ChangedAt:    history.ChangedAt,
			ChangeReason: history.ChangeReason,
		})
	}

	return response, nil
}
