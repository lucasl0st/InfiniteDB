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

	var rows []struct {
		database                 string
		table                    string
		objectsInsertedPerSecond int64
		totalObjects             int64
	}

	for database, databaseMetric := range databaseMetrics {
		for tableName, tableMetric := range databaseMetric.Tables {
			rows = append(rows, struct {
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

	sort.Slice(rows, func(i, j int) bool {
		return rows[i].totalObjects > rows[j].totalObjects
	})

	for _, row := range rows {
		t.AppendRow(table.Row{
			row.database,
			row.table,
			row.objectsInsertedPerSecond,
			row.totalObjects,
		})
	}

	t.Render()
}

func renderPerformanceMetrics() {
	t := table.NewWriter()
	t.SetOutputMirror(tm.Output)

	t.AppendHeader(table.Row{"Function Name", "Average Call Time"})

	var rows []struct {
		functionName    string
		averageCallTime time.Duration
	}

	for function, performanceMetric := range performanceMetrics.Functions {
		rows = append(rows, struct {
			functionName    string
			averageCallTime time.Duration
		}{functionName: function, averageCallTime: performanceMetric.AverageFunctionCallDuration})
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i].averageCallTime > rows[j].averageCallTime
	})

	for _, row := range rows {
		t.AppendRow(table.Row{
			row.functionName,
			fmt.Sprint(row.averageCallTime),
		})
	}

	t.Render()
}
