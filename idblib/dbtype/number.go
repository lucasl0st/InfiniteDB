/*
 * Copyright (c) 2023 Lucas Pape
 */

package dbtype

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"regexp"
)

type Number struct {
	n    big.Float
	null bool
}

func NumberFromString(s string) (Number, error) {
	if len(s) == 0 {
		return NumberFromNull(), nil
	}

	p := uint(math.Ceil(math.Log2(10)*float64(len(s)))) + 1
	b, ok := new(big.Float).SetPrec(p).SetString(s)

	if !ok {
		return Number{}, errors.New("could not convert string into number")
	}

	return Number{n: *b, null: false}, nil
}

func NumberFromInt64(i int64) (Number, error) {
	return NumberFromString(fmt.Sprint(i))
}

func NumberFromFloat64(f float64) (Number, error) {
	return Number{
		n:    *new(big.Float).SetFloat64(f),
		null: false,
	}, nil
}

func NumberFromNull() Number {
	return Number{
		n:    *big.NewFloat(0),
		null: true,
	}
}

func (a Number) Larger(b DBType) bool {
	bn := b.(Number)
	return a.n.Cmp(&bn.n) > 0 && !(a.null || bn.null)
}

func (a Number) Smaller(b DBType) bool {
	bn := b.(Number)
	return a.n.Cmp(&bn.n) < 0 && !(a.null || bn.null)
}

func (a Number) Equal(b DBType) bool {
	bn := b.(Number)
	return a.n.Cmp(&bn.n) == 0 || ((a.null || bn.null) && (a.null == bn.null))
}

func (a Number) Not(b DBType) bool {
	bn := b.(Number)
	return a.n.Cmp(&bn.n) != 0 || ((a.null || bn.null) && (a.null != bn.null))
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

	return a.n.Text('f', -1)
}

func (a Number) ToJsonRaw() json.RawMessage {
	return json.RawMessage(a.ToString())
}

func (a Number) IsNull() bool {
	return a.null
}

func (a Number) ToFloat64() float64 {
	f, _ := a.n.Float64()

	return f
}
