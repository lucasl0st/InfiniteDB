/*
 * Copyright (c) 2023 Lucas Pape
 */

package dbtype

import (
	"encoding/json"
	"fmt"
	"regexp"
)

type Text struct {
	s    string
	null bool
}

func TextFromString(s string) Text {
	return Text{s: s, null: false}
}

func TextFromNull() Text {
	return Text{
		s:    "",
		null: true,
	}
}

func (a Text) Larger(b DBType) bool {
	bt := b.(Text)
	return a.s > bt.s && !(a.null || bt.null)
}

func (a Text) Smaller(b DBType) bool {
	bt := b.(Text)
	return a.s < bt.s && !(a.null || bt.null)
}

func (a Text) Equal(b DBType) bool {
	bt := b.(Text)
	return a.s == bt.s || ((a.null || bt.null) && (a.null == bt.null))
}

func (a Text) Not(b DBType) bool {
	bt := b.(Text)
	return a.s != bt.s || ((a.null || bt.null) && (a.null != bt.null))
}

func (a Text) Matches(r regexp.Regexp) bool {
	return r.MatchString(a.s)
}

func (a Text) Between(s DBType, l DBType) bool {
	return a.Larger(s) && a.Smaller(l)
}

func (a Text) ToString() string {
	if a.null {
		return "null"
	}

	return a.s
}

func (a Text) ToJsonRaw() json.RawMessage {
	return json.RawMessage(fmt.Sprintf("\"%s\"", a.ToString()))
}

func (a Text) IsNull() bool {
	return a.null
}
