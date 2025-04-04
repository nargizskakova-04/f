package server

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"

	v1 "frappuccino/internal/delivery/http/v1"
	serviceInv "frappuccino/internal/service/inventory"
	serviceMenu "frappuccino/internal/service/menu"
	serviceOrder "frappuccino/internal/service/order"
	serviceReport "frappuccino/internal/service/report"

	"frappuccino/internal/config"
	"frappuccino/internal/repository/postgres"
)

type App struct {
	cfg    *config.Config
	router *http.ServeMux
	logger *log.Logger
	db     *sql.DB
}

func NewApp(cfg *config.Config) *App {
	return &App{cfg: cfg}
}

func (app *App) Initialize() error {
	logger := log.New(os.Stdout, "http: ", log.LstdFlags)
	app.logger = logger
	app.router = http.NewServeMux()
	if err := app.setHandler(); err != nil {
		app.logger.Println("method:Initialize, function:setHandler", err.Error())
		return err
	}
	return nil
}

func (app *App) Run() {
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.cfg.App.Port),
		Handler:      app.router,
		ReadTimeout:  app.cfg.App.RTO,
		WriteTimeout: app.cfg.App.WTO,
	}
	app.logger.Println("Starting server on port", app.cfg.App.Port)
	if err := server.ListenAndServe(); err != nil {
		app.logger.Println("server shutdown: ", err)
		return
	}
}

func (app *App) setHandler() error {
	var err error

	/*dbConn*/
	dbConn, err := postgres.NewDbConnInstance(&app.cfg.Repository)

	inventoryRepository := postgres.NewInventoryRepository(dbConn)
	inventoryService := serviceInv.NewInventoryService(inventoryRepository, app.logger)

	v1.SetInventoryHandler(app.router, inventoryService, app.logger)
	if err != nil {
		app.logger.Println("Connection to db failed")
		return err
	}

	menuRepository := postgres.NewMenuRepository(dbConn)
	menuService := serviceMenu.NewMenuService(menuRepository, app.logger)

	v1.SetMenuHandler(app.router, menuService, app.logger)
	if err != nil {
		app.logger.Println("Connection to db failed")
		return err
	}

	orderRepository := postgres.NewOrderRepository(dbConn)
	orderService := serviceOrder.NewOrderService(
		orderRepository,
		menuRepository,      // Required for ingredient checks
		inventoryRepository, // Required for inventory updates
		app.logger,
	)

	v1.SetOrderHandler(app.router, orderService, app.logger)
	if err != nil {
		app.logger.Println("Connection to db failed")
		return err
	}

	// Add search repository and service
	searchRepository := postgres.NewSearchRepository(dbConn)
	searchService := serviceReport.NewSearchService(searchRepository, app.logger)

	// Add report handler
	v1.SetReportHandler(app.router, searchService, app.logger)

	return nil
}
