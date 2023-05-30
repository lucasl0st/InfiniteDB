/*
 * Copyright (c) 2023 Lucas Pape
 */

package client

import (
	"fmt"
	"github.com/lucasl0st/InfiniteDB/models/metric"
)

func (c *Client) handleMetricsUpdateMethod(msg map[string]interface{}) {
	if c.MetricsReceiver != nil {
		m := msg["metric"]

		switch m {
		case fmt.Sprint(metric.DatabaseMetric):
			var databaseMetricsResponse metric.DatabaseMetricResponse

			err := mapToStruct(msg["value"].(map[string]interface{}), &databaseMetricsResponse)

			if err != nil {
				panic(err.Error())
			}

			c.MetricsReceiver.DatabaseMetrics(databaseMetricsResponse.Database, databaseMetricsResponse.Metrics)

		case fmt.Sprint(metric.PerformanceMetric):
			var performanceMetricsResponse metric.PerformanceMetricResponse

			err := mapToStruct(msg["value"].(map[string]interface{}), &performanceMetricsResponse)

			if err != nil {
				panic(err.Error())
			}

			c.MetricsReceiver.PerformanceMetrics(performanceMetricsResponse.Metrics)
		case fmt.Sprint(metric.MemStatsMetric):
			var memStatsMetricResponse metric.MemStatsMetricResponse

			err := mapToStruct(msg["value"].(map[string]interface{}), &memStatsMetricResponse)

			if err != nil {
				panic(err.Error())
			}

			c.MetricsReceiver.MemStatsMetrics(memStatsMetricResponse.Metrics)
		}
	}
}
