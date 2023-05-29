/*
 * Copyright (c) 2023 Lucas Pape
 */

package metrics

import (
	"github.com/lucasl0st/InfiniteDB/models/metric"
	"sync"
	"time"
)

type Metrics struct {
	databasesLock sync.RWMutex
	databases     map[string]metric.DatabaseMetrics

	r *metric.Receiver
}

func New(receiver *metric.Receiver) *Metrics {
	m := Metrics{
		databases: map[string]metric.DatabaseMetrics{},

		r: receiver,
	}

	go func() {
		for {
			m.runner()

			time.Sleep(time.Second)
		}
	}()

	return &m
}

func (m *Metrics) createMetrics(database string, table string) {
	m.databasesLock.Lock()
	defer m.databasesLock.Unlock()

	_, ok := m.databases[database]

	if !ok {
		m.databases[database] = metric.DatabaseMetrics{Tables: map[string]metric.TableMetrics{}}
	}

	_, ok = m.databases[database].Tables[table]

	if !ok {
		m.databases[database].Tables[table] = metric.TableMetrics{
			InsertedObjects: 0,
			TotalObjects:    0,
		}
	}
}

func (m *Metrics) WroteObject(database string, table string) {
	m.createMetrics(database, table)

	m.databasesLock.Lock()
	defer m.databasesLock.Unlock()

	tableMetric := m.databases[database].Tables[table]
	tableMetric.InsertedObjects += 1
	m.databases[database].Tables[table] = tableMetric
}

func (m *Metrics) AddTotalObject(database string, table string) {
	m.createMetrics(database, table)

	m.databasesLock.Lock()
	defer m.databasesLock.Unlock()

	tableMetric := m.databases[database].Tables[table]
	tableMetric.TotalObjects += 1
	m.databases[database].Tables[table] = tableMetric
}

func (m *Metrics) runner() {
	m.databasesLock.Lock()
	defer m.databasesLock.Unlock()

	for database, databaseMetrics := range m.databases {
		(*m.r).DatabaseMetrics(database, databaseMetrics)

		for table, tableMetrics := range databaseMetrics.Tables {
			tableMetrics.InsertedObjects = 0

			databaseMetrics.Tables[table] = tableMetrics
		}

		m.databases[database] = databaseMetrics
	}

	averageTimingMeasurements := getAverageTimingMeasurements()

	performanceMetrics := metric.PerformanceMetrics{Functions: map[string]metric.FunctionMetrics{}}

	for name, average := range averageTimingMeasurements {
		performanceMetrics.Functions[name] = metric.FunctionMetrics{AverageFunctionCallDuration: average}
	}

	(*m.r).PerformanceMetrics(performanceMetrics)
}
