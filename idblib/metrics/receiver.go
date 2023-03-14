/*
 * Copyright (c) 2023 Lucas Pape
 */

package metrics

type Receiver interface {
	ObjectsInsertedPerSecond(v int64)
	TotalObjects(v int64)
}
