/*
 * Copyright (c) 2023 Lucas Pape
 */

package field

import (
	"github.com/lucasl0st/InfiniteDB/idblib/object"
	"github.com/lucasl0st/InfiniteDB/util"
	"testing"
)

func generateTestData() (map[string]Field, []object.Object) {
	numberOfObjects := 1000

	testFields := map[string]Field{
		"textField": {
			Name:    "textField",
			Indexed: true,
			Unique:  false,
			Null:    false,
			Type:    TEXT,
		},
		"numberField": {
			Name:    "numberField",
			Indexed: true,
			Unique:  false,
			Null:    false,
			Type:    NUMBER,
		},
		"booleanField": {
			Name:    "booleanField",
			Indexed: true,
			Unique:  false,
			Null:    false,
			Type:    BOOL,
		},
	}

	var objects []object.Object

	for i := 0; i < numberOfObjects; i++ {
		o := object.Object{
			Id: int64(i),
			M: map[string]interface{}{
				"textField":    util.RandomString(32),
				"numberField":  util.RandomFloat(),
				"booleanField": util.RandomBoolean(),
			},
		}

		objects = append(objects, o)
	}

	return testFields, objects
}

func expectObjects(t *testing.T, v []int64, e []int64) {
	if len(v) != len(e) {
		t.Errorf("expected length is %v but received length is %v", len(e), len(v))
	}

	for _, i := range v {
		found := false

		for _, ei := range e {
			if i == ei {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("could not find expected value")
		}
	}
}

func TestIndex_GetValue(t *testing.T) {
	testFields, testObjects := generateTestData()

	i := NewIndex(testFields)

	for _, testObject := range testObjects {
		i.Index(testObject)
	}

	for fieldName, field := range testFields {
		if !field.Indexed {
			continue
		}

		for _, testObject := range testObjects {
			v := i.GetValue(fieldName, testObject.Id)

			if v != util.InterfaceToString(testObject.M[fieldName]) {
				t.Errorf("expected value %v, received value %s", testObject.M[fieldName], v)
			}
		}
	}
}

func TestIndex_Equal(t *testing.T) {
	testFields, testObjects := generateTestData()

	i := NewIndex(testFields)

	for _, testObject := range testObjects {
		i.Index(testObject)
	}

	for fieldName, field := range testFields {
		if !field.Indexed {
			continue
		}

		for _, testObject := range testObjects {
			var expected []int64

			for _, not := range testObjects {
				if util.InterfaceToString(not.M[fieldName]) == util.InterfaceToString(testObject.M[fieldName]) {
					expected = append(expected, not.Id)
				}
			}

			v := i.Equal(fieldName, util.InterfaceToString(testObject.M[fieldName]))

			expectObjects(t, v, expected)
		}
	}
}

func TestIndex_Not(t *testing.T) {
	testFields, testObjects := generateTestData()

	i := NewIndex(testFields)

	for _, testObject := range testObjects {
		i.Index(testObject)
	}

	for fieldName, field := range testFields {
		if !field.Indexed {
			continue
		}

		for _, testObject := range testObjects {
			var expected []int64

			for _, not := range testObjects {
				if util.InterfaceToString(not.M[fieldName]) != util.InterfaceToString(testObject.M[fieldName]) {
					expected = append(expected, not.Id)
				}
			}

			v := i.Not(fieldName, util.InterfaceToString(testObject.M[fieldName]))

			expectObjects(t, v, expected)
		}
	}
}

func TestIndex_Larger(t *testing.T) {
	testFields, testObjects := generateTestData()

	i := NewIndex(testFields)

	for _, testObject := range testObjects {
		i.Index(testObject)
	}

	for fieldName, field := range testFields {
		if !field.Indexed {
			continue
		}

		for _, testObject := range testObjects {
			var expected []int64

			for _, larger := range testObjects {
				switch field.Type {
				case TEXT:
					if larger.M[fieldName].(string) > testObject.M[fieldName].(string) {
						expected = append(expected, larger.Id)
					}
				case NUMBER:
					if larger.M[fieldName].(float64) > testObject.M[fieldName].(float64) {
						expected = append(expected, larger.Id)
					}
				case BOOL:
					if util.InterfaceToString(larger.M[fieldName]) > util.InterfaceToString(testObject.M[fieldName].(bool)) {
						expected = append(expected, larger.Id)
					}
				}
			}

			v, err := i.Larger(fieldName, util.InterfaceToString(testObject.M[fieldName]), field.Type == NUMBER)

			if err != nil {
				t.Error(err.Error())
			}

			expectObjects(t, v, expected)
		}
	}
}

func TestIndex_Smaller(t *testing.T) {
	testFields, testObjects := generateTestData()

	i := NewIndex(testFields)

	for _, testObject := range testObjects {
		i.Index(testObject)
	}

	for fieldName, field := range testFields {
		if !field.Indexed {
			continue
		}

		for _, testObject := range testObjects {
			var expected []int64

			for _, larger := range testObjects {
				switch field.Type {
				case TEXT:
					if larger.M[fieldName].(string) < testObject.M[fieldName].(string) {
						expected = append(expected, larger.Id)
					}
				case NUMBER:
					if larger.M[fieldName].(float64) < testObject.M[fieldName].(float64) {
						expected = append(expected, larger.Id)
					}
				case BOOL:
					if util.InterfaceToString(larger.M[fieldName]) < util.InterfaceToString(testObject.M[fieldName].(bool)) {
						expected = append(expected, larger.Id)
					}
				}
			}

			v, err := i.Smaller(fieldName, util.InterfaceToString(testObject.M[fieldName]), field.Type == NUMBER)

			if err != nil {
				t.Error(err.Error())
			}

			expectObjects(t, v, expected)
		}
	}
}
