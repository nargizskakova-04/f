package v1

import (
	"net/http"
)

func setInventoryRoutes(handler *InventoryHandler, router *http.ServeMux) {
	router.HandleFunc("POST /inventory", handler.CreateInventoryRequest)
	router.HandleFunc("GET /inventory", handler.GetInventoryResponse)
	router.HandleFunc("GET /inventory/{id}", handler.GetInventoryByIDResponse) // New endpoint
	router.HandleFunc("DELETE /inventory/{id}", handler.DeleteInventoryRequest)
	router.HandleFunc("PUT /inventory/{id}", handler.UpdateInventoryRequest)
	router.HandleFunc("POST /inventory/transactions", handler.CreateInventoryTransactionRequest)
	router.HandleFunc("GET /inventory/{id}/transactions", handler.GetInventoryTransactionsResponse)
	router.HandleFunc("GET /inventory/getLeftOvers", handler.GetLeftOversResponse)
}

func setMenuRoutes(handler *MenuHandler, router *http.ServeMux) {
	router.HandleFunc("POST /menu", handler.CreateMenuItemRequest)
	router.HandleFunc("GET /menu", handler.GetMenuResponse)
	router.HandleFunc("GET /menu/{id}", handler.GetMenuByIDResponse) // New endpoint
	router.HandleFunc("DELETE /menu/{id}", handler.DeleteMenuRequest)
	router.HandleFunc("PUT /menu/{id}", handler.UpdateMenuRequest)
	router.HandleFunc("GET /price-history", handler.GetAllPriceHistoryResponse)

}

func setOrderRoutes(handler *OrderHandler, router *http.ServeMux) {
	// router.HandleFunc("GET /orders/", handler.GetOrderByID)
	router.HandleFunc("GET /orders/{id}", handler.GetOrderByIDResponse)
	router.HandleFunc("GET /orders", handler.GetOrderResponse)
	router.HandleFunc("POST /orders", handler.CreateOrderRequest)
	router.HandleFunc("PUT /orders/{id}", handler.UpdateOrderRequest)
	router.HandleFunc("GET /order-status-history", handler.GetAllOrderStatusHistory)
	router.HandleFunc("DELETE /orders/{id}", handler.DeleteOrderRequest)
	router.HandleFunc("POST /orders/{id}/close", handler.CloseOrder)
	router.HandleFunc("GET /orders/numberOfOrderedItems", handler.GetNumberOfOrderedItems)
	router.HandleFunc("POST /orders/batch-process", handler.BatchProcessOrdersRequest)
}

func setReportRoutes(handler *ReportHandler, router *http.ServeMux) {
	router.HandleFunc("GET /reports/search", handler.SearchReport)
	router.HandleFunc("GET /reports/orderedItemsByPeriod", handler.GetOrderedItemsByPeriod)
}

// func SetOrderHandler(router *http.ServeMux, orderService order.ServiceInterface, logger *log.Logger) {
// 	handler := NewOrderHandler(orderService)
// 	setOrderRoutes(handler, router)
// }
