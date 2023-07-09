/*
 * Copyright (c) 2023 Lucas Pape
 */

package metric

import (
	"runtime"
	"time"
)

type Metric string

const DatabaseMetric Metric = "databaseMetric"
const PerformanceMetric Metric = "performanceMetric"
const MemStatsMetric Metric = "memStatMetric"

type DatabaseMetrics struct {
	Tables map[string]TableMetrics `json:"tables"`
}

type PerformanceMetrics struct {
	Functions map[string]FunctionMetrics `json:"functions"`
}

type FunctionMetrics struct {
	AverageFunctionCallDuration time.Duration `json:"averageFunctionCallDuration"`
}

type TableMetrics struct {
	InsertedObjects int64 `json:"insertedObjects"`
	TotalObjects    int64 `json:"totalObjects"`
}

type MemStatsMetrics struct {
	MemStats runtime.MemStats `json:"memStats"`
}
