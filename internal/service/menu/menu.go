package menu

import (
	"context"
	"errors"
	"log"
	"time"

	"frappuccino/internal/dto/menu"
	"frappuccino/internal/entity"
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

	id, err := s.menuRepo.CreateMenuItem(ctx, insertToDBMenuItem)
	if err != nil {
		s.logger.Println("CreateMenuItem error:", err)
		return "", err
	}

	var ingredients []entity.MenuItemIngredient
	for _, dtoIng := range request.Ingredients {
		ingredients = append(ingredients, entity.MenuItemIngredient{
			MenuItemID:   id,
			IngredientID: dtoIng.IngredientID,
			Quantity:     dtoIng.Quantity,
			Unit:         dtoIng.Unit,
		})
	}

	if err := s.menuRepo.CreateMenuItemIngredients(ctx, id, ingredients); err != nil {
		s.logger.Println("CreateMenuItemIngredients error:", err)
		return "", err
	}

	return id, nil
}

func (s *MenuService) GetMenuItem(ctx context.Context) ([]menu.GetMenuResponse, error) {
	items, err := s.menuRepo.GetMenuItem(ctx)
	if err != nil {
		s.logger.Println("Error retrieving menu items:", err)
		return nil, err
	}

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
	updates := make(map[string]interface{})

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
	updates["updated_at"] = time.Now()

	if len(updates) == 1 && updates["updated_at"] != nil {
		return "", errors.New("no fields to update")
	}

	id, err := s.menuRepo.UpdateMenu(ctx, updates, id)
	if err != nil {
		s.logger.Println(err)
		return "", err
	}

	return id, nil
}

func (s *MenuService) GetAllPriceHistory(ctx context.Context) ([]menu.GetPriceHistoryResponse, error) {
	histories, err := s.menuRepo.GetAllPriceHistory(ctx)
	if err != nil {
		s.logger.Println("Error retrieving price history:", err)
		return nil, err
	}

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
