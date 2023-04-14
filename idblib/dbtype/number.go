/*
 * Copyright (c) 2023 Lucas Pape
 */

package dbtype

import (
	"fmt"
	"regexp"
	"strconv"
)

type Number struct {
	n    float64
	null bool
}

func NumberFromString(s string) (Number, error) {
	f, err := strconv.ParseFloat(s, 64)

	if err != nil {
		return Number{}, err
	}

	return Number{n: f, null: false}, nil
}

func NumberFromFloat64(f float64) Number {
	return Number{n: f, null: false}
}

func NumberFromNull() Number {
	return Number{
		n:    0,
		null: true,
	}
}

func (a Number) Larger(b DBType) bool {
	bn := b.(Number)
	return a.n > bn.n && !(a.null || bn.null)
}

func (a Number) Smaller(b DBType) bool {
	bn := b.(Number)
	return a.n < bn.n && !(a.null || bn.null)
}

func (a Number) Equal(b DBType) bool {
	bn := b.(Number)
	return a.n == bn.n || ((a.null || bn.null) && (a.null == bn.null))
}

func (a Number) Not(b DBType) bool {
	bn := b.(Number)
	return a.n != bn.n || ((a.null || bn.null) && (a.null != bn.null))
}

func (a Number) Matches(r regexp.Regexp) bool {
	return r.MatchString(a.ToString())
}

func (a Number) Between(s DBType, l DBType) bool {
	return a.Larger(s) && a.Smaller(l)
}

func (a Number) ToString() string {
	if a.null {
		return "null"
	}

	return fmt.Sprint(a.n)
}

func (a Number) IsNull() bool {
	return a.null
}

func (a Number) ToFloat64() float64 {
	return a.n
}
