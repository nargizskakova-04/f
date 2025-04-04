package report

import "time"

// OrderedItemsByPeriodRequest represents the parameters for retrieving ordered items by period
type OrderedItemsByPeriodRequest struct {
	Period string    // "day" or "month"
	Month  string    // Required when Period="day"
	Year   string    // Used for both period types, defaults to current year if not specified
	Start  time.Time // Calculated from Month and Year
	End    time.Time // Calculated from Month and Year
}

// DayCount represents the count of ordered items for a specific day
type DayCount struct {
	Day   int `json:"day"`
	Count int `json:"count"`
}

// MonthCount represents the count of ordered items for a specific month
type MonthCount struct {
	Month string `json:"month"`
	Count int    `json:"count"`
}

// OrderedItemsByPeriodResponse represents the response for ordered items by period
type OrderedItemsByPeriodResponse struct {
	Period       string      `json:"period"`
	Month        string      `json:"month,omitempty"`
	Year         string      `json:"year,omitempty"`
	OrderedItems interface{} `json:"orderedItems"`
}
