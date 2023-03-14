package util

import (
	"fmt"
	"reflect"
	"strconv"
)

func InterfaceToString(i interface{}) string {
	if i == nil {
		return "null"
	}

	if reflect.ValueOf(i).Kind() == reflect.Ptr {
		return toStringFromPtr(i)
	}

	return toString(i)
}

func StringToNumber(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

func toString(i interface{}) string {
	b, isBoolean := i.(bool)

	if isBoolean {
		if b {
			return "true"
		} else {
			return "false"
		}
	}

	f, isFloat := i.(float64)

	if isFloat {
		return strconv.FormatFloat(f, 'f', -1, 64)
	}

	s, isString := i.(string)

	if isString {
		return s
	}

	return fmt.Sprint(i)
}

func toStringFromPtr(i interface{}) string {
	b, isBoolean := i.(*bool)

	if isBoolean {
		if *b {
			return "true"
		} else {
			return "false"
		}
	}

	f, isFloat := i.(*float64)

	if isFloat {
		return strconv.FormatFloat(*f, 'f', -1, 64)
	}

	s, isString := i.(*string)

	if isString {
		return *s
	}

	return fmt.Sprint(i)
}
