/*
 * Copyright (c) 2023 Lucas Pape
 */

package methods

import (
	"fmt"
	tm "github.com/buger/goterm"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/lucasl0st/InfiniteDB/client"
	"github.com/lucasl0st/InfiniteDB/models/metric"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"
)

func init() {
	Methods = append(Methods, Method{
		Name:      "metrics",
		Arguments: []Argument{},
		Run:       runMetrics,
	})
}

var renderingMetrics = false

var databaseMetrics map[string]metric.DatabaseMetrics
var performanceMetrics metric.PerformanceMetrics

type MetricsReceiver struct {
}

func (r MetricsReceiver) DatabaseMetrics(database string, m metric.DatabaseMetrics) {
	databaseMetrics[database] = m

	renderMetrics()
}

func (r MetricsReceiver) PerformanceMetrics(m metric.PerformanceMetrics) {
	performanceMetrics = m

	renderMetrics()
}

func runMetrics(c *client.Client, _ []string) error {
	databaseMetrics = map[string]metric.DatabaseMetrics{}

	renderingMetrics = true

	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch

		renderingMetrics = false

		signal.Stop(ch)
	}()

	c.MetricsReceiver = MetricsReceiver{}

	_, err := c.SubscribeToMetricUpdates()

	if err != nil {
		return err
	}

	renderMetrics()

	for renderingMetrics {
		time.Sleep(time.Millisecond * 100)
	}

	c.MetricsReceiver = nil

	_, err = c.UnsubscribeFromMetricUpdates()

	tm.Clear()
	tm.Flush()

	if err != nil {
		return err
	}

	return nil
}

func renderMetrics() {
	if !renderingMetrics {
		return
	}

	tm.Clear()
	tm.MoveCursor(1, 1)

	renderDatabaseMetrics()
	renderPerformanceMetrics()

	tm.Flush()
}

func renderDatabaseMetrics() {
	t := table.NewWriter()
	t.SetOutputMirror(tm.Output)

	t.AppendHeader(table.Row{"Database", "Table", "Objects Inserted Per Second", "Total Objects"})

	var values []struct {
		database                 string
		table                    string
		objectsInsertedPerSecond int64
		totalObjects             int64
	}

	for database, databaseMetric := range databaseMetrics {
		for tableName, tableMetric := range databaseMetric.Tables {
			values = append(values, struct {
				database                 string
				table                    string
				objectsInsertedPerSecond int64
				totalObjects             int64
			}{
				database:                 database,
				table:                    tableName,
				objectsInsertedPerSecond: tableMetric.InsertedObjects,
				totalObjects:             tableMetric.TotalObjects,
			})
		}
	}

	sort.Slice(values, func(i, j int) bool {
		return values[i].totalObjects > values[j].totalObjects
	})

	for _, value := range values {
		t.AppendRow(table.Row{
			value.database,
			value.table,
			value.objectsInsertedPerSecond,
			value.totalObjects,
		})
	}

	t.Render()
}

func renderPerformanceMetrics() {
	t := table.NewWriter()
	t.SetOutputMirror(tm.Output)

	t.AppendHeader(table.Row{"Average Insert Time", "Average Get Time"})
	t.AppendRow(table.Row{
		fmt.Sprint(performanceMetrics.AverageObjectInsertTime),
		fmt.Sprint(performanceMetrics.AverageObjectGetTime),
	})

	t.Render()
}
