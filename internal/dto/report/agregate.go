package report

import "time"

// TotalSalesRequest represents the request parameters for total sales report
type TotalSalesRequest struct {
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	Status    string     `json:"status,omitempty"` // Filter by order status (e.g., "delivered")
}

// TotalSalesResponse represents the response for total sales report
type TotalSalesResponse struct {
	TotalSales       float64   `json:"total_sales"`
	OrderCount       int       `json:"order_count"`
	AverageOrderSize float64   `json:"average_order_size"`
	StartDate        time.Time `json:"start_date,omitempty"`
	EndDate          time.Time `json:"end_date,omitempty"`
	Status           string    `json:"status,omitempty"`
}

// PopularItemsRequest represents the request parameters for popular items report
type PopularItemsRequest struct {
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	Limit     int        `json:"limit,omitempty"` // Number of items to return
}

// PopularItem represents a single popular menu item with its sales data
type PopularItem struct {
	MenuItemID     string  `json:"menu_item_id"`
	Name           string  `json:"name"`
	Quantity       int     `json:"quantity_sold"`
	TotalRevenue   float64 `json:"total_revenue"`
	PercentOfSales float64 `json:"percent_of_sales,omitempty"`
}

// PopularItemsResponse represents the response for popular items report
type PopularItemsResponse struct {
	Items         []PopularItem `json:"items"`
	TotalQuantity int           `json:"total_quantity"`
	TotalRevenue  float64       `json:"total_revenue"`
	StartDate     time.Time     `json:"start_date,omitempty"`
	EndDate       time.Time     `json:"end_date,omitempty"`
	MostPopular   string        `json:"most_popular"`
	MostRevenue   string        `json:"most_revenue"`
}
