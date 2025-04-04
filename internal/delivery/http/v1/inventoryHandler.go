package v1

import (
	"log"
	"net/http"
)

type InventoryHandler struct {
	logger           *log.Logger
	inventoryService inventoryInterface
}

func NewInventoryHandler(
	inventoryService inventoryInterface,
	logger *log.Logger,
) *InventoryHandler {
	return &InventoryHandler{
		inventoryService: inventoryService,
		logger:           logger,
	}
}

func SetInventoryHandler(
	router *http.ServeMux,
	inventoryService inventoryInterface,
	logger *log.Logger,
) {
	handler := NewInventoryHandler(inventoryService, logger)
	setInventoryRoutes(handler, router)
}
