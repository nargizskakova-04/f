package inventory

import (
	"context"
	"errors"
	"log"
	"time"

	"frappuccino/internal/dto/inventory"
	"frappuccino/internal/entity"
)

type InventoryService struct {
	inventoryRepo inventoryRepo
	logger        *log.Logger
}

func NewInventoryService(inventoryRepo inventoryRepo, logger *log.Logger) *InventoryService {
	return &InventoryService{
		inventoryRepo: inventoryRepo,
		logger:        logger,
	}
}

func (s *InventoryService) CreateInventory(ctx context.Context, request inventory.CreateInventoryRequest) (string, error) {
	insertToDBInventopory := entity.Inventory{
		Name:         request.Name,
		Quantity:     request.Quantity,
		Unit:         request.Unit,
		UnitPrice:    request.UnitPrice,
		LastUpdated:  time.Now(),
		ReorderPoint: request.ReorderPoint,
	}
	id, err := s.inventoryRepo.CreateInventory(ctx, insertToDBInventopory)
	if err != nil {
		s.logger.Println(err)
		return "", err
	}

	if request.Quantity > 0 {
		transaction := entity.InventoryTransaction{
			IngredientID:    id,
			QuantityChange:  request.Quantity,
			TransactionType: "addition",
			Reason:          "Initial inventory creation",
		}

		if err := s.inventoryRepo.CreateInventoryTransaction(ctx, transaction); err != nil {
			s.logger.Println("Failed to record inventory transaction:", err)
		}
	}
	return id, nil
}

func (s *InventoryService) GetInventory(ctx context.Context) ([]inventory.GetInventoryResponse, error) {
	items, err := s.inventoryRepo.GetInventory(ctx)
	if err != nil {
		s.logger.Println("Error retrieving inventory items:", err)
		return nil, err
	}

	var response []inventory.GetInventoryResponse
	for _, item := range items {
		response = append(response, inventory.GetInventoryResponse{
			IngredientID: item.IngredientID,
			Name:         item.Name,
			Quantity:     item.Quantity,
			Unit:         item.Unit,
			UnitPrice:    item.UnitPrice,
			LastUpdated:  item.LastUpdated,
			ReorderPoint: item.ReorderPoint,
		})
	}

	return response, nil
}

func (s *InventoryService) GetInventoryByID(ctx context.Context, id string) (inventory.GetInventoryResponse, error) {
	item, err := s.inventoryRepo.GetInventoryByID(ctx, id)
	if err != nil {
		s.logger.Println("Error retrieving inventory item:", err)
		return inventory.GetInventoryResponse{}, err
	}

	response := inventory.GetInventoryResponse{
		IngredientID: item.IngredientID,
		Name:         item.Name,
		Quantity:     item.Quantity,
		Unit:         item.Unit,
		UnitPrice:    item.UnitPrice,
		LastUpdated:  item.LastUpdated,
		ReorderPoint: item.ReorderPoint,
	}

	return response, nil
}

func (s *InventoryService) DeleteInventory(ctx context.Context, id string) (string, error) {
	inventory, err := s.inventoryRepo.GetInventoryByID(ctx, id)
	if err != nil {
		s.logger.Println("Error retrieving inventory item:", err)
		return "", err
	}

	if inventory.Quantity > 0 {
		transaction := entity.InventoryTransaction{
			IngredientID:    id,
			QuantityChange:  inventory.Quantity,
			TransactionType: "deduction",
			Reason:          "Inventory item deleted",
		}

		if err := s.inventoryRepo.CreateInventoryTransaction(ctx, transaction); err != nil {
			s.logger.Println("Failed to record inventory transaction for deletion:", err)
		}
	}

	ingredient_id, err := s.inventoryRepo.DeleteInventory(ctx, id)
	if err != nil {
		s.logger.Println(err)
		return "", err
	}

	return ingredient_id, nil
}

func (s *InventoryService) UpdateInventory(ctx context.Context, request inventory.UpdateInventoryRequest, id string) (string, error) {
	currentInventory, err := s.inventoryRepo.GetInventoryByID(ctx, id)
	if err != nil {
		s.logger.Println("Error retrieving current inventory:", err)
		return "", err
	}

	updates := make(map[string]interface{})

	if request.Name != nil {
		updates["name"] = *request.Name
	}

	var quantityChanged bool
	var oldQuantity float32
	var newQuantity float32

	if request.Quantity != nil {
		oldQuantity = currentInventory.Quantity
		newQuantity = *request.Quantity
		updates["quantity"] = newQuantity
		quantityChanged = true
	}

	if request.Unit != nil {
		updates["unit"] = *request.Unit
	}

	if request.UnitPrice != nil {
		updates["unit_price"] = *request.UnitPrice
	}

	if request.ReorderPoint != nil {
		updates["reorder_point"] = *request.ReorderPoint
	}

	updates["last_updated"] = time.Now()

	if len(updates) == 1 && updates["last_updated"] != nil {
		return "", errors.New("no fields to update")
	}

	id, err = s.inventoryRepo.UpdateInventory(ctx, updates, id)
	if err != nil {
		s.logger.Println(err)
		return "", err
	}

	if quantityChanged {
		quantityDifference := newQuantity - oldQuantity
		transactionType := "addition"
		if quantityDifference < 0 {
			transactionType = "deduction"
			quantityDifference = -quantityDifference
		} else if quantityDifference == 0 {
			return id, nil
		}

		transaction := entity.InventoryTransaction{
			IngredientID:    id,
			QuantityChange:  quantityDifference,
			TransactionType: transactionType,
			Reason:          "Inventory update",
		}

		if err := s.inventoryRepo.CreateInventoryTransaction(ctx, transaction); err != nil {
			s.logger.Println("Failed to record inventory transaction:", err)
		}
	}

	return id, nil
}

func (s *InventoryService) RecordInventoryTransaction(ctx context.Context, request inventory.CreateTransactionRequest) error {
	_, err := s.inventoryRepo.GetInventoryByID(ctx, request.IngredientID)
	if err != nil {
		return errors.New("ingredient not found")
	}

	transaction := entity.InventoryTransaction{
		IngredientID:    request.IngredientID,
		QuantityChange:  request.QuantityChange,
		TransactionType: request.TransactionType,
		Reason:          request.Reason,
	}

	if err := s.inventoryRepo.CreateInventoryTransaction(ctx, transaction); err != nil {
		return err
	}

	currentInventory, err := s.inventoryRepo.GetInventoryByID(ctx, request.IngredientID)
	if err != nil {
		return err
	}

	var newQuantity float32
	switch request.TransactionType {
	case "addition":
		newQuantity = currentInventory.Quantity + request.QuantityChange
	case "deduction":
		newQuantity = currentInventory.Quantity - request.QuantityChange
		if newQuantity < 0 {
			return errors.New("insufficient inventory")
		}
	case "adjustment":
		newQuantity = request.QuantityChange
	case "waste":
		newQuantity = currentInventory.Quantity - request.QuantityChange
		if newQuantity < 0 {
			return errors.New("insufficient inventory")
		}
	default:
		return errors.New("invalid transaction type")
	}

	updates := map[string]interface{}{
		"quantity":     newQuantity,
		"last_updated": time.Now(),
	}

	_, err = s.inventoryRepo.UpdateInventory(ctx, updates, request.IngredientID)
	return err
}

func (s *InventoryService) GetInventoryTransactions(ctx context.Context, ingredientID string) ([]inventory.TransactionResponse, error) {
	_, err := s.inventoryRepo.GetInventoryByID(ctx, ingredientID)
	if err != nil {
		return nil, errors.New("ingredient not found")
	}

	transactions, err := s.inventoryRepo.GetInventoryTransactions(ctx, ingredientID)
	if err != nil {
		s.logger.Println("Error retrieving inventory transactions:", err)
		return nil, err
	}

	var response []inventory.TransactionResponse
	for _, tx := range transactions {
		response = append(response, inventory.TransactionResponse{
			TransactionID:   tx.TransactionID,
			IngredientID:    tx.IngredientID,
			QuantityChange:  tx.QuantityChange,
			TransactionType: tx.TransactionType,
			Reason:          tx.Reason,
			CreatedAt:       tx.CreatedAt,
		})
	}

	return response, nil
}

func (s *InventoryService) GetLeftOvers(ctx context.Context, sortBy string, page, pageSize int) (inventory.GetLeftOversResponse, error) {
	items, totalCount, err := s.inventoryRepo.GetLeftOvers(ctx, sortBy, page, pageSize)
	if err != nil {
		s.logger.Println("Error retrieving inventory leftovers:", err)
		return inventory.GetLeftOversResponse{}, err
	}

	totalPages := (totalCount + pageSize - 1) / pageSize
	hasNextPage := page < totalPages

	var leftoverItems []inventory.LeftOverItem
	for _, item := range items {
		leftoverItems = append(leftoverItems, inventory.LeftOverItem{
			Name:     item.Name,
			Quantity: item.Quantity,
			Price:    item.UnitPrice * 100,
		})
	}

	return inventory.GetLeftOversResponse{
		CurrentPage: page,
		HasNextPage: hasNextPage,
		PageSize:    pageSize,
		TotalPages:  totalPages,
		Data:        leftoverItems,
	}, nil
}
