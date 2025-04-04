package v1

import (
	"log"
	"net/http"
)

// ReportHandler handles report-related operations
type ReportHandler struct {
	logger        *log.Logger
	reportService reportInterface
}

// NewReportHandler creates a new report handler
func NewReportHandler(
	reportService reportInterface,
	logger *log.Logger,
) *ReportHandler {
	return &ReportHandler{
		reportService: reportService,
		logger:        logger,
	}
}

// SetReportHandler sets up report-related routes
func SetReportHandler(
	router *http.ServeMux,
	reportService reportInterface,
	logger *log.Logger,
) {
	handler := NewReportHandler(reportService, logger)
	setReportRoutes(handler, router)
}
