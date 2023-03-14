/*
 * Copyright (c) 2023 Lucas Pape
 */

package functions

import (
	e "github.com/lucasl0st/InfiniteDB/errors"
	"github.com/lucasl0st/InfiniteDB/idblib/object"
	"github.com/lucasl0st/InfiniteDB/idblib/table"
	"github.com/lucasl0st/InfiniteDB/util"
	"strings"
)

const fieldNameMath = "math"

type MathFunction struct {
	formula string
	as      string
}

func (m *MathFunction) Run(
	table *table.Table,
	objects object.Objects,
	additionalFields table.AdditionalFields,
	parameters map[string]interface{},
) (object.Objects, table.AdditionalFields, error) {
	err := m.parseParameters(parameters)

	if err != nil {
		return nil, nil, err
	}

	for _, o := range objects {
		result, err := m.runFormula(table, o, additionalFields)

		if err != nil {
			return nil, nil, err
		}

		if additionalFields[o] == nil {
			additionalFields[o] = make(map[string]interface{})
		}

		additionalFields[o][m.as] = result
	}

	return objects, additionalFields, nil
}

func (m *MathFunction) parseParameters(parameters map[string]interface{}) error {
	f, ok := parameters["formula"].(string)

	if !ok {
		return e.IsNotAString("formula in math function")
	}

	as, ok := parameters["as"].(string)

	if !ok {
		as = fieldNameMath
	}

	m.formula = f
	m.as = as

	return nil
}

func (m *MathFunction) runFormula(table *table.Table, object int64, additionalFields table.AdditionalFields) (float64, error) {
	words := strings.Split(m.formula, " ")

	var err error
	first := true
	var result float64 = 0

	operand := ""

	for _, word := range words {
		var workingValue float64 = 0

		switch word {
		case "+":
			operand = "+"
		case "-":
			operand = "-"
		case "*":
			operand = ":"
		case "/":
			operand = "/"
		default:
			if strings.HasPrefix(word, "$") {
				fieldName := trimFirstRune(word)

				if additionalFields[object][fieldName] != nil {
					workingValue = additionalFields[object][fieldName].(float64)
				} else {
					workingValue, err = util.StringToNumber(table.Index.GetValue(fieldName, object))

					if err != nil {
						return 0, err
					}
				}
			} else {
				f, err := util.StringToNumber(word)

				if err != nil {
					return 0, err
				}

				workingValue = f
			}

			if first {
				result = workingValue
				first = false
				continue
			} else if len(operand) == 0 {
				return 0, e.CouldNotParseFormula()
			}

			switch operand {
			case "+":
				result += workingValue
			case "-":
				result -= workingValue
			case "*":
				result *= workingValue
			case "/":
				if workingValue == 0 {
					return 0, e.CannotDivideByZero()
				}

				result /= workingValue
			}
		}
	}

	return result, nil
}

func trimFirstRune(s string) string {
	for i := range s {
		if i > 0 {
			// The value i is the index in s of the second
			// rune.  Slice to remove the first rune.
			return s[i:]
		}
	}
	// There are 0 or 1 runes in the string.
	return ""
}
