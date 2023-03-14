/*
 * Copyright (c) 2023 Lucas Pape
 */

package util

import (
	"fmt"
	"testing"
)

func TestInterfaceToString(t *testing.T) {
	v := InterfaceToString("test")
	tv := "test"

	if v != tv {
		t.Errorf("expected %s, received %v", tv, v)
	}

	v = InterfaceToString(nil)
	tv = "null"

	if v != tv {
		t.Errorf("expected %s, received %v", tv, v)
	}

	v = InterfaceToString(420.0)
	tv = "420"

	if v != tv {
		t.Errorf("expected %s, received %v", tv, v)
	}

	f := 420.0
	v = InterfaceToString(&f)
	tv = "420"

	if v != tv {
		t.Errorf("expected %s, received %v", tv, v)
	}

	s := "hello"
	v = InterfaceToString(&s)
	tv = "hello"

	if v != tv {
		t.Errorf("expected %s, received %v", tv, v)
	}

	v = InterfaceToString(true)
	tv = "true"

	if v != tv {
		t.Errorf("expected %s, received %v", tv, v)
	}

	v = InterfaceToString(false)
	tv = "false"

	if v != tv {
		t.Errorf("expected %s, received %v", tv, v)
	}

	b := false
	v = InterfaceToString(&b)
	tv = "false"

	if v != tv {
		t.Errorf("expected %s, received %v", tv, v)
	}

	b = true
	v = InterfaceToString(&b)
	tv = "true"

	if v != tv {
		t.Errorf("expected %s, received %v", tv, v)
	}

	v = InterfaceToString(10)
	tv = "10"

	if v != tv {
		t.Errorf("expected %s, received %v", tv, v)
	}

	i := 20
	p := &i
	v = InterfaceToString(p)
	tv = fmt.Sprint(p)

	if v != tv {
		t.Errorf("expected %s, received %v", tv, v)
	}
}

func TestStringToNumber(t *testing.T) {
	n := "123456789"
	f := 123456789.0

	v, err := StringToNumber(n)

	if err != nil {
		t.Error(err.Error())
	}

	if v != f {
		t.Errorf("expected %v, received %v", f, v)
	}
}
