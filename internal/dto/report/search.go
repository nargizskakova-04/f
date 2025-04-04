package report

// SearchRequest represents the parameters for a search operation
type SearchRequest struct {
	Query    string  `json:"q"`
	Filter   string  `json:"filter,omitempty"`
	MinPrice float64 `json:"min_price,omitempty"`
	MaxPrice float64 `json:"max_price,omitempty"`
}

// SearchResultItem represents a menu item in search results
type SearchResultMenuItem struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Price       float32 `json:"price"`
	Relevance   float64 `json:"relevance"`
}

// SearchResultOrder represents an order in search results
type SearchResultOrder struct {
	ID           string   `json:"id"`
	CustomerName string   `json:"customer_name"`
	Items        []string `json:"items"`
	Total        float64  `json:"total"`
	Relevance    float64  `json:"relevance"`
}

// SearchResponse represents the combined search results
type SearchResponse struct {
	MenuItems    []SearchResultMenuItem `json:"menu_items,omitempty"`
	Orders       []SearchResultOrder    `json:"orders,omitempty"`
	TotalMatches int                    `json:"total_matches"`
}
