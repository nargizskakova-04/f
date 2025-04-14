package report

import "time"

type OrderedItemsByPeriodRequest struct {
	Period string
	Month  string
	Year   string
	Start  time.Time
	End    time.Time
}

type DayCount struct {
	Day   int `json:"day"`
	Count int `json:"count"`
}

type MonthCount struct {
	Month string `json:"month"`
	Count int    `json:"count"`
}

type OrderedItemsByPeriodResponse struct {
	Period       string      `json:"period"`
	Month        string      `json:"month,omitempty"`
	Year         string      `json:"year,omitempty"`
	OrderedItems interface{} `json:"orderedItems"`
}
