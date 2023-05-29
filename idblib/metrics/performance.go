/*
 * Copyright (c) 2023 Lucas Pape
 */

package metrics

import (
	"runtime"
	"sync"
	"time"
)

type Measurement struct {
	Start time.Time
	End   time.Time
}

var measurements map[string]map[int64]Measurement
var measurementsLock sync.RWMutex
var id int64 = 0

func init() {
	measurements = map[string]map[int64]Measurement{}
}

func getFunctionName() string {
	pc, _, _, _ := runtime.Caller(2)
	details := runtime.FuncForPC(pc)
	name := details.Name()

	return name
}

func createMapEntry(name string) {
	_, ok := measurements[name]

	if !ok {
		measurements[name] = map[int64]Measurement{}
	}
}

func StartTimingMeasurement() int64 {
	name := getFunctionName()

	measurementsLock.Lock()
	defer measurementsLock.Unlock()

	createMapEntry(name)

	id++
	measurementId := id

	measurements[name][measurementId] = Measurement{
		Start: time.Now(),
	}

	return measurementId
}

func StopTimingMeasurement(id int64) {
	end := time.Now()

	name := getFunctionName()

	measurementsLock.Lock()
	defer measurementsLock.Unlock()

	m := measurements[name][id]
	m.End = end
	measurements[name][id] = m
}

func getAverageTimingMeasurements() map[string]time.Duration {
	measurementsLock.RLock()
	defer measurementsLock.RUnlock()

	m := map[string]time.Duration{}

	for name, measurements := range measurements {
		var count int64 = 0

		for _, measurement := range measurements {
			if measurement.End.Nanosecond() != 0 {
				m[name] += measurement.End.Sub(measurement.Start)
				count++
			}
		}

		if count != 0 {
			m[name] = time.Duration(m[name].Nanoseconds() / count)
		}
	}

	return m
}
