/*
 * Copyright (c) 2023 Lucas Pape
 */

package table

import (
	e "github.com/lucasl0st/InfiniteDB/errors"
	"github.com/lucasl0st/InfiniteDB/idblib/field"
	"github.com/lucasl0st/InfiniteDB/idblib/metrics"
	"github.com/lucasl0st/InfiniteDB/idblib/object"
	"github.com/lucasl0st/InfiniteDB/idblib/storage"
	"github.com/lucasl0st/InfiniteDB/request"
	"github.com/lucasl0st/InfiniteDB/util"
	"os"
	"regexp"
	gsort "sort"
	"strings"
)

type Table struct {
	DatabaseName string
	Name         string
	path         string
	Config       field.TableConfig
	Index        *field.Index
	Storage      *storage.Storage
}

func NewTable(
	databaseName string,
	name string,
	path string,
	config field.TableConfig,
	logger util.Logger,
	metrics *metrics.Metrics,
	cacheSize uint,
) (*Table, error) {
	table := Table{
		DatabaseName: databaseName,
		Name:         name,
		path:         path,
		Config:       config,
		Index:        field.NewIndex(config.Fields),
	}

	s, err := storage.NewStorage(path+name+"/", table.addedObject, table.deletedObject, logger, metrics, cacheSize)

	if err != nil {
		return nil, err
	}

	table.Storage = s

	return &table, err
}

func (t *Table) Delete() error {
	t.Storage.Kill()

	return os.RemoveAll(t.path + t.Name)
}

func (t *Table) addedObject(object object.Object) {
	t.Index.Index(object)
}

func (t *Table) deletedObject(object object.Object) {
	t.Index.UnIndex(object)
}

func (t *Table) Where(w request.Where, andObjects object.Objects) (object.Objects, error) {
	v := util.InterfaceToString(w.Value)

	parseNumber := t.Config.Fields[w.Field].Type == field.NUMBER

	var err error
	var results object.Objects

	switch w.Operator {
	case request.EQUALS:
		if andObjects == nil {
			results = t.Index.Equal(w.Field, v)
		} else {
			results = t.andEqual(andObjects, w.Field, v)
		}
	case request.NOT:
		if andObjects == nil {
			results = t.Index.Not(w.Field, v)
		} else {
			results = t.andNot(andObjects, w.Field, v)
		}
	case request.MATCH:
		r, err := regexp.Compile(v)

		if err != nil {
			return nil, err
		}

		if andObjects == nil {
			results = t.Index.Match(w.Field, *r)
		} else {
			results = t.andMatch(andObjects, w.Field, *r)
		}
	case request.SMALLER:
		if andObjects == nil {
			results, err = t.Index.Smaller(w.Field, v, parseNumber)
		} else {
			results, err = t.andSmaller(andObjects, w.Field, v, parseNumber)
		}
	case request.LARGER:
		if andObjects == nil {
			results, err = t.Index.Larger(w.Field, v, parseNumber)
		} else {
			results, err = t.andLarger(andObjects, w.Field, v, parseNumber)
		}
	case request.BETWEEN:
		values := strings.Split(v, "-")

		if len(values) <= 1 {
			return nil, e.NotEnoughValuesForOperator(w.Operator)
		}

		if andObjects == nil {
			results, err = t.Index.Between(w.Field, values[0], values[1], parseNumber)
		} else {
			results, err = t.andBetween(andObjects, w.Field, values[0], values[1], parseNumber)
		}
	default:
		return nil, e.NotAValidOperator()
	}

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (t *Table) Query(q Query, andObjects object.Objects, additionalFields AdditionalFields) (object.Objects, AdditionalFields, error) {
	runMiddleware, runQuery := QueryMiddleware(t, q)

	if runMiddleware {
		return runQuery(andObjects)
	}

	var objects object.Objects
	var err error

	if q.Where != nil {
		if q.Where.All != nil && len(q.Where.All) > 0 {
			query := Query{
				Where: &request.Where{
					Field:    q.Where.Field,
					Operator: q.Where.Operator,
					Value:    q.Where.All[0],
				},
			}

			nextQuery := &query

			for _, a := range q.Where.All {
				if a == q.Where.All[0] {
					continue
				}

				nextQuery.And = &Query{
					Where: &request.Where{
						Field:    q.Where.Field,
						Operator: q.Where.Operator,
						Value:    a,
					},
				}

				nextQuery = nextQuery.And
			}

			objects, additionalFields, err = t.Query(query, andObjects, additionalFields)
		} else if q.Where.Any != nil && len(q.Where.Any) > 0 {
			query := Query{
				Where: &request.Where{
					Field:    q.Where.Field,
					Operator: q.Where.Operator,
					Value:    q.Where.Any[0],
				},
			}

			nextQuery := &query

			for _, a := range q.Where.Any {
				if a == q.Where.Any[0] {
					continue
				}

				nextQuery.Or = &Query{
					Where: &request.Where{
						Field:    q.Where.Field,
						Operator: q.Where.Operator,
						Value:    a,
					},
				}

				nextQuery = nextQuery.Or
			}

			objects, additionalFields, err = t.Query(query, andObjects, additionalFields)
		} else {
			objects, err = t.Where(*q.Where, andObjects)
		}

		if err != nil {
			return nil, nil, err
		}
	}

	if q.Functions != nil {
		for _, function := range q.Functions {
			objects, additionalFields, err = function.Function.Run(t, objects, additionalFields, function.Parameters)

			if err != nil {
				return nil, nil, err
			}
		}
	}

	if q.And != nil {
		objects, additionalFields, err = t.Query(*q.And, objects, additionalFields)

		if err != nil {
			return nil, nil, err
		}
	}

	if q.Or != nil {
		var next object.Objects

		next, additionalFields, err = t.Query(*q.Or, andObjects, additionalFields)

		if err != nil {
			return nil, additionalFields, err
		}

		if next != nil {
			objects = append(objects, next...)
			objects = t.removeDuplicates(objects)
		}
	}

	return objects, additionalFields, nil
}

func (t *Table) Insert(objectM map[string]interface{}) error {
	o := object.Object{
		M: objectM,
	}

	runMiddleware, insert := InsertMiddleware(t, &o)

	if runMiddleware {
		return insert()
	}

	err := t.allFieldsHaveValues(&o)

	if err != nil {
		return err
	}

	err = t.isUnique(&o)

	if err != nil {
		return err
	}

	t.Storage.AddObject(o)

	if err != nil {
		return err
	}

	return nil
}

func (t *Table) Update(object *object.Object) error {
	runMiddleware, update := UpdateMiddleware(t, object)

	if runMiddleware {
		return update()
	}

	err := t.allFieldsHaveValues(object)

	if err != nil {
		return err
	}

	err = t.isUnique(object)

	if err != nil {
		return err
	}

	t.Storage.AddObject(*object)

	if err != nil {
		return err
	}

	t.Index.UpdateIndex(*object)

	return nil
}

func (t *Table) Remove(object *object.Object) error {
	runMiddleware, remove := RemoveMiddleware(t, object)

	if runMiddleware {
		return remove()
	}

	t.Index.UnIndex(*object)
	t.Storage.DeleteObject(*object)

	return nil
}

func (t *Table) FindExisting(object map[string]interface{}) (int64, error) {
	for fieldName, f := range t.Config.Fields {
		if f.Indexed && f.Unique {
			indexElements := t.Index.Equal(fieldName, util.InterfaceToString(object[fieldName]))

			if len(indexElements) > 0 {
				return indexElements[0], nil
			}
		}
	}

	return 0, e.CouldNotFindObjectWithAtLeastOneIndexedAndUniqueValue()
}

func (t *Table) and(objects object.Objects, otherObjects object.Objects) object.Objects {
	results := object.Objects{}

	for _, o := range objects {
		for _, oo := range otherObjects {
			if o == oo {
				results = append(results, o)
				break
			}
		}
	}

	return results
}

func (t *Table) Sort(o object.Objects, fieldName string, additionalFields AdditionalFields, direction request.SortDirection) (object.Objects, error) {
	f := t.Config.Fields[fieldName]

	switch f.Type {
	case field.TEXT:
		return t.sortString(o, fieldName, direction, nil), nil
	case field.NUMBER:
		return t.sortFloat(o, fieldName, direction, nil), nil
	case field.BOOL:
		return t.sortBoolean(o, fieldName, direction, nil), nil
	default:
		if len(o) > 0 {
			sampleValue := additionalFields[o[0]][fieldName]

			_, isText := sampleValue.(string)

			if isText {
				return t.sortString(o, fieldName, direction, additionalFields), nil
			}

			_, isNumber := sampleValue.(float64)

			if isNumber {
				return t.sortFloat(o, fieldName, direction, additionalFields), nil
			}

			_, isBoolean := sampleValue.(bool)

			if isBoolean {
				return t.sortBoolean(o, fieldName, direction, additionalFields), nil
			}
		} else {
			return o, nil
		}

		return o, e.CannotSortType()
	}
}

func (t *Table) sortString(o object.Objects, fieldName string, direction request.SortDirection, additionalFields AdditionalFields) object.Objects {
	switch direction {
	case request.ASC:
		gsort.Slice(o, func(i, j int) bool {
			iv := ""
			jv := ""

			if additionalFields != nil {
				iv = additionalFields[o[i]][fieldName].(string)
				jv = additionalFields[o[j]][fieldName].(string)
			} else {
				iv = t.Index.GetValue(fieldName, o[i])
				jv = t.Index.GetValue(fieldName, o[j])
			}

			return iv < jv
		})
	case request.DESC:
		gsort.Slice(o, func(i, j int) bool {
			iv := ""
			jv := ""

			if additionalFields != nil {
				iv = additionalFields[o[i]][fieldName].(string)
				jv = additionalFields[o[j]][fieldName].(string)
			} else {
				iv = t.Index.GetValue(fieldName, o[i])
				jv = t.Index.GetValue(fieldName, o[j])
			}

			return iv > jv
		})
	}

	return o
}

func (t *Table) sortFloat(o object.Objects, fieldName string, direction request.SortDirection, additionalFields AdditionalFields) object.Objects {
	switch direction {
	case request.ASC:
		gsort.Slice(o, func(i, j int) bool {
			var iv float64 = 0
			var jv float64 = 0

			if additionalFields != nil {
				iv = additionalFields[o[i]][fieldName].(float64)
				jv = additionalFields[o[j]][fieldName].(float64)
			} else {
				iv, _ = util.StringToNumber(t.Index.GetValue(fieldName, o[i]))
				jv, _ = util.StringToNumber(t.Index.GetValue(fieldName, o[j]))
			}

			return iv < jv
		})
	case request.DESC:
		gsort.Slice(o, func(i, j int) bool {
			var iv float64 = 0
			var jv float64 = 0

			if additionalFields != nil {
				iv = additionalFields[o[i]][fieldName].(float64)
				jv = additionalFields[o[j]][fieldName].(float64)
			} else {
				iv, _ = util.StringToNumber(t.Index.GetValue(fieldName, o[i]))
				jv, _ = util.StringToNumber(t.Index.GetValue(fieldName, o[j]))
			}

			return iv > jv
		})
	}

	return o
}

func (t *Table) sortBoolean(o object.Objects, fieldName string, direction request.SortDirection, additionalFields AdditionalFields) object.Objects {
	switch direction {
	case request.ASC:
		gsort.Slice(o, func(i, j int) bool {
			iv := false

			if additionalFields != nil {
				iv = additionalFields[o[i]][fieldName].(bool)
			} else {
				iv = t.Index.GetValue(fieldName, o[i]) == "true"
			}

			return iv
		})
	case request.DESC:
		gsort.Slice(o, func(i, j int) bool {
			iv := false

			if additionalFields != nil {
				iv = additionalFields[o[i]][fieldName].(bool)
			} else {
				iv = t.Index.GetValue(fieldName, o[i]) == "true"
			}

			return !iv
		})
	}

	return o
}

func (t *Table) SkipAndLimit(objects object.Objects, skip *int64, limit *int64) object.Objects {
	var results object.Objects

	if skip != nil && limit != nil {
		for i, o := range objects {
			if int64(i) >= *skip && int64(i) < (*skip+*limit) {
				results = append(results, o)
			}
		}

	} else if skip == nil && limit != nil {
		for i, o := range objects {
			if int64(i) < *limit {
				results = append(results, o)
			}
		}

	} else if skip != nil && limit == nil {
		for i, o := range objects {
			if int64(i) >= *skip {
				results = append(results, o)
			}
		}
	} else if skip == nil && limit == nil {
		results = objects
	}

	return results
}

func (t *Table) isUnique(o *object.Object) error {
	for fieldName, f := range t.Config.Fields {
		if f.Unique {
			if len(t.Index.Equal(fieldName, util.InterfaceToString(o.M[fieldName]))) > 0 {
				return e.FoundExistingObjectWithField(fieldName)
			}
		}
	}

	for _, fieldNames := range t.Config.Options.CombinedUniques {
		var first = true
		var objects object.Objects

		for _, fieldName := range fieldNames {
			if first {
				objects = t.Index.Equal(fieldName, util.InterfaceToString(o.M[fieldName]))
				first = false
			} else {
				objects = t.and(objects, t.Index.Equal(fieldName, util.InterfaceToString(o.M[fieldName])))

				if len(objects) == 0 {
					break
				}
			}
		}

		if len(objects) != 0 {
			return e.FoundExistingObjectWithCombinedUniques()
		}
	}

	return nil
}

func (t *Table) allFieldsHaveValues(o *object.Object) error {
	for fieldName, f := range t.Config.Fields {
		if o.M[fieldName] == nil && !f.Null {
			return e.ObjectDoesNotHaveValueForField(fieldName)
		}
	}

	return nil
}

func (t *Table) andEqual(andObjects object.Objects, field string, value string) object.Objects {
	results := object.Objects{}

	for _, andObject := range andObjects {
		if t.Index.GetValue(field, andObject) == value {
			results = append(results, andObject)
		}
	}

	return results
}

func (t *Table) andNot(andObjects object.Objects, field string, value string) object.Objects {
	results := object.Objects{}

	for _, andObject := range andObjects {
		if t.Index.GetValue(field, andObject) != value {
			results = append(results, andObject)
		}
	}

	return results
}

func (t *Table) andMatch(andObjects object.Objects, field string, r regexp.Regexp) object.Objects {
	results := object.Objects{}

	for _, andObject := range andObjects {
		if r.MatchString(t.Index.GetValue(field, andObject)) {
			results = append(results, andObject)
		}
	}

	return results
}

func (t *Table) andSmaller(andObjects object.Objects, field string, value string, parseNumber bool) (object.Objects, error) {
	results := object.Objects{}

	for _, andObject := range andObjects {
		if parseNumber {
			valueInt, err := util.StringToNumber(value)

			if err != nil {
				return nil, err
			}

			andObjectValueInt, err := util.StringToNumber(t.Index.GetValue(field, andObject))

			if err != nil {
				return nil, err
			}

			if andObjectValueInt < valueInt {
				results = append(results, andObject)
			}
		} else {
			if t.Index.GetValue(field, andObject) < value {
				results = append(results, andObject)
			}
		}
	}

	return results, nil
}

func (t *Table) andLarger(andObjects object.Objects, field string, value string, parseNumber bool) (object.Objects, error) {
	results := object.Objects{}

	for _, andObject := range andObjects {
		if parseNumber {
			valueInt, err := util.StringToNumber(value)

			if err != nil {
				return nil, err
			}

			andObjectValueInt, err := util.StringToNumber(t.Index.GetValue(field, andObject))

			if err != nil {
				return nil, err
			}

			if andObjectValueInt > valueInt {
				results = append(results, andObject)
			}
		} else {
			if t.Index.GetValue(field, andObject) > value {
				results = append(results, andObject)
			}
		}
	}

	return results, nil
}

func (t *Table) andBetween(andObjects object.Objects, field string, smaller string, larger string, parseNumber bool) (object.Objects, error) {
	results := object.Objects{}

	for _, andObject := range andObjects {
		if parseNumber {
			smallerInt, err := util.StringToNumber(smaller)

			if err != nil {
				return nil, err
			}

			largerInt, err := util.StringToNumber(larger)

			if err != nil {
				return nil, err
			}

			andObjectValueInt, err := util.StringToNumber(t.Index.GetValue(field, andObject))

			if err != nil {
				return nil, err
			}

			if andObjectValueInt > smallerInt && andObjectValueInt < largerInt {
				results = append(results, andObject)
			}
		} else {
			andObjectValue := t.Index.GetValue(field, andObject)

			if andObjectValue > smaller && andObjectValue < larger {
				results = append(results, andObject)
			}
		}
	}

	return results, nil
}

func (t *Table) removeDuplicates(o object.Objects) object.Objects {
	duplicates := map[int64]bool{}
	var results object.Objects

	for _, o := range o {
		_, duplicate := duplicates[o]

		if !duplicate {
			duplicates[o] = true
			results = append(results, o)
		}
	}

	return results
}
