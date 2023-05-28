/*
 * Copyright (c) 2023 Lucas Pape
 */

package metric

type DatabaseMetricResponse struct {
	Database string          `json:"database"`
	Metrics  DatabaseMetrics `json:"metrics"`
}

type PerformanceMetricResponse struct {
	Metrics PerformanceMetrics `json:"metrics"`
}
