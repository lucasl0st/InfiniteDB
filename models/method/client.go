/*
 * Copyright (c) 2023 Lucas Pape
 */

package method

type ClientMethod string

const HeloMethod ClientMethod = "helo"
const GenericErrorMethod ClientMethod = "genericError"
const RequestResponseMethod ClientMethod = "requestResponse"
const MetricsUpdateMethod ClientMethod = "metricsUpdate"
