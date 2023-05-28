/*
 * Copyright (c) 2023 Lucas Pape
 */

package metric

type Receiver interface {
	DatabaseMetrics(database string, m DatabaseMetrics)
	PerformanceMetrics(m PerformanceMetrics)
}
