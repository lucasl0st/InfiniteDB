/*
 * Copyright (c) 2023 Lucas Pape
 */

package metrics

import (
	"time"
)

type Metrics struct {
	lastReportedTotalObjects int64

	insertedObjects int64
	totalObjects    int64

	r *Receiver
}

func New(receiver *Receiver) *Metrics {
	m := Metrics{
		lastReportedTotalObjects: 0,

		insertedObjects: 0,
		totalObjects:    0,

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

func (m *Metrics) WroteObject() {
	m.insertedObjects += 1
}

func (m *Metrics) AddTotalObject() {
	m.totalObjects += 1
}

func (m *Metrics) runner() {
	objectsPerSecond := m.insertedObjects
	m.insertedObjects = 0

	if objectsPerSecond != 0 {
		if m.r != nil {
			(*m.r).ObjectsInsertedPerSecond(objectsPerSecond)
		}
	}

	if m.lastReportedTotalObjects != m.totalObjects {
		m.lastReportedTotalObjects = m.totalObjects

		if m.r != nil {
			(*m.r).TotalObjects(m.totalObjects)
		}
	}
}
