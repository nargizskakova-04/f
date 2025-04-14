package report

import "time"

type TotalSalesRequest struct {
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	Status    string     `json:"status,omitempty"`
}

type TotalSalesResponse struct {
	TotalSales       float64   `json:"total_sales"`
	OrderCount       int       `json:"order_count"`
	AverageOrderSize float64   `json:"average_order_size"`
	StartDate        time.Time `json:"start_date,omitempty"`
	EndDate          time.Time `json:"end_date,omitempty"`
	Status           string    `json:"status,omitempty"`
}

type PopularItemsRequest struct {
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	Limit     int        `json:"limit,omitempty"`
}

type PopularItem struct {
	MenuItemID     string  `json:"menu_item_id"`
	Name           string  `json:"name"`
	Quantity       int     `json:"quantity_sold"`
	TotalRevenue   float64 `json:"total_revenue"`
	PercentOfSales float64 `json:"percent_of_sales,omitempty"`
}

type PopularItemsResponse struct {
	Items         []PopularItem `json:"items"`
	TotalQuantity int           `json:"total_quantity"`
	TotalRevenue  float64       `json:"total_revenue"`
	StartDate     time.Time     `json:"start_date,omitempty"`
	EndDate       time.Time     `json:"end_date,omitempty"`
	MostPopular   string        `json:"most_popular"`
	MostRevenue   string        `json:"most_revenue"`
}
