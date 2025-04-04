package v1

import (
	"log"
	"net/http"
)

type MenuHandler struct {
	logger      *log.Logger
	menuService menuInterface
}

func NewMenuHandler(
	menuService menuInterface,
	logger *log.Logger,
) *MenuHandler {
	return &MenuHandler{
		menuService: menuService,
		logger:      logger,
	}
}

func SetMenuHandler(
	router *http.ServeMux,
	menuService menuInterface,
	logger *log.Logger,
) {
	handler := NewMenuHandler(menuService, logger)
	setMenuRoutes(handler, router)
}
