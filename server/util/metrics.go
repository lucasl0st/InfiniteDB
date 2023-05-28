/*
 * Copyright (c) 2023 Lucas Pape
 */

package util

import "github.com/lucasl0st/InfiniteDB/models/metric"

type MetricsReceiver struct {
	SubmitMetric func(metric metric.Metric, value any)
}

func (r *MetricsReceiver) DatabaseMetrics(database string, m metric.DatabaseMetrics) {
	r.SubmitMetric(metric.DatabaseMetric, metric.DatabaseMetricResponse{
		Database: database,
		Metrics:  m,
	})
}

func (r *MetricsReceiver) PerformanceMetrics(m metric.PerformanceMetrics) {
	r.SubmitMetric(metric.PerformanceMetric, metric.PerformanceMetricResponse{Metrics: m})
}
