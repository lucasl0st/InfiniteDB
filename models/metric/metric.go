/*
 * Copyright (c) 2023 Lucas Pape
 */

package metric

import "time"

type Metric string

const DatabaseMetric Metric = "databaseMetric"
const PerformanceMetric Metric = "performanceMetric"

type DatabaseMetrics struct {
	Tables map[string]TableMetrics `json:"tables"`
}

type PerformanceMetrics struct {
	AverageObjectInsertTime time.Duration `json:"averageInsertTime"`
	AverageObjectGetTime    time.Duration `json:"averageObjectGetTime"`
}

type TableMetrics struct {
	InsertedObjects int64 `json:"insertedObjects"`
	TotalObjects    int64 `json:"totalObjects"`
}
