/*
 * Copyright (c) 2023 Lucas Pape
 */

package dbtype

import (
	"regexp"
)

type Bool struct {
	b    bool
	null bool
}

func BoolFromString(s string) Bool {
	if s == "true" {
		return Bool{b: true, null: false}
	} else {
		return Bool{b: false, null: false}
	}
}

func BoolFromBool(b bool) Bool {
	return Bool{b: b, null: false}
}

func BoolFromNull() Bool {
	return Bool{
		b:    false,
		null: true,
	}
}

func (a Bool) Larger(b DBType) bool {
	bb := b.(Bool)

	return a.b && !(a.null || bb.null)
}

func (a Bool) Smaller(b DBType) bool {
	bb := b.(Bool)
	return !a.b && !(a.null || bb.null)
}

func (a Bool) Equal(b DBType) bool {
	bb := b.(Bool)
	return a.b == bb.b || ((a.null || bb.null) && (a.null == bb.null))
}

func (a Bool) Not(b DBType) bool {
	bb := b.(Bool)
	return a.b != bb.b || ((a.null || bb.null) && (a.null != bb.null))
}

func (a Bool) Matches(r regexp.Regexp) bool {
	return r.MatchString(a.ToString())
}

func (a Bool) Between(s DBType, l DBType) bool {
	return a.Larger(s) && a.Smaller(l)
}

func (a Bool) ToString() string {
	if a.null {
		return "null"
	}

	if a.b {
		return "true"
	} else {
		return "false"
	}
}

func (a Bool) IsNull() bool {
	return a.null
}

func (a Bool) ToBool() bool {
	return a.b
}
