package v1

import (
	"log"
	"net/http"
)

type ReportHandler struct {
	logger        *log.Logger
	reportService reportInterface
}

func NewReportHandler(
	reportService reportInterface,
	logger *log.Logger,
) *ReportHandler {
	return &ReportHandler{
		reportService: reportService,
		logger:        logger,
	}
}

func SetReportHandler(
	router *http.ServeMux,
	reportService reportInterface,
	logger *log.Logger,
) {
	handler := NewReportHandler(reportService, logger)
	setReportRoutes(handler, router)
}
