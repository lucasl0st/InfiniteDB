/*
 * Copyright (c) 2023 Lucas Pape
 */

package dbtype

import (
	"encoding/json"
	"regexp"
)

type DBType interface {
	Larger(b DBType) bool
	Smaller(b DBType) bool
	Equal(b DBType) bool
	Not(b DBType) bool
	Matches(r regexp.Regexp) bool
	Between(s DBType, l DBType) bool
	ToString() string
	ToJsonRaw() json.RawMessage
	IsNull() bool
}
