package v1

import (
	"log"
	"net/http"
)

type OrderHandler struct {
	logger       *log.Logger
	orderService orderInterface
}

func NewOrderHandler(
	orderService orderInterface,
	logger *log.Logger,
) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
		logger:       logger,
	}
}

func SetOrderHandler(
	router *http.ServeMux,
	orderService orderInterface,
	logger *log.Logger,
) {
	handler := NewOrderHandler(orderService, logger)
	setOrderRoutes(handler, router)
}
