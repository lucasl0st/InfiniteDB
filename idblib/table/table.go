/*
 * Copyright (c) 2023 Lucas Pape
 */

package table

import (
	"fmt"
	e "github.com/lucasl0st/InfiniteDB/errors"
	"github.com/lucasl0st/InfiniteDB/idblib/dbtype"
	"github.com/lucasl0st/InfiniteDB/idblib/field"
	"github.com/lucasl0st/InfiniteDB/idblib/metrics"
	"github.com/lucasl0st/InfiniteDB/idblib/object"
	"github.com/lucasl0st/InfiniteDB/idblib/storage"
	idbutil "github.com/lucasl0st/InfiniteDB/idblib/util"
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

	s, err := storage.NewStorage(path+name+"/", table.addedObject, table.deletedObject, logger, metrics, cacheSize, config.Fields)

	if err != nil {
		return nil, err
	}

	table.Storage = s

	table.Config.Fields[field.InternalObjectIdField] = field.Field{
		Name:    field.InternalObjectIdField,
		Indexed: true,
		Unique:  true,
		Null:    false,
		Type:    field.NUMBER,
	}

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
	f := t.Config.Fields[w.Field]

	var results object.Objects

	switch w.Operator {
	case request.MATCH:
		s, ok := w.Value.(string)

		if !ok {
			return nil, e.ValueForOperatorMustBeString(request.MATCH)
		}

		r, err := regexp.Compile(s)

		if err != nil {
			return nil, err
		}

		if andObjects == nil {
			results = t.Index.Match(w.Field, *r)
		} else {
			results = t.andMatch(andObjects, w.Field, *r)
		}
	case request.BETWEEN:
		s, ok := w.Value.(string)

		if !ok {
			return nil, e.ValueForOperatorMustBeString(request.BETWEEN)
		}

		values := strings.Split(s, "_")

		if len(values) <= 1 {
			return nil, e.NotEnoughValuesForOperator(w.Operator)
		}

		smaller, err := idbutil.StringToDBType(values[0], f)

		if err != nil {
			return nil, err
		}

		larger, err := idbutil.StringToDBType(values[1], f)

		if andObjects == nil {
			results, err = t.Index.Between(w.Field, smaller, larger)
		} else {
			results, err = t.andBetween(andObjects, w.Field, smaller, larger)
		}

		if err != nil {
			return nil, err
		}
	default:
		v, err := idbutil.InterfaceToDBType(w.Value, f)

		if err != nil {
			return nil, err
		}

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
		case request.SMALLER:
			if andObjects == nil {
				results, err = t.Index.Smaller(w.Field, v)
			} else {
				results, err = t.andSmaller(andObjects, w.Field, v)
			}
		case request.LARGER:
			if andObjects == nil {
				results, err = t.Index.Larger(w.Field, v)
			} else {
				results, err = t.andLarger(andObjects, w.Field, v)
			}
		default:
			return nil, e.NotAValidOperator()
		}
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
	runMiddleware, insert := InsertMiddleware(t, objectM)

	if runMiddleware {
		return insert()
	}

	o, err := t.InterfaceMapToObject(objectM)

	if err != nil {
		return err
	}

	err = t.allFieldsHaveValues(o)

	if err != nil {
		return err
	}

	err = t.isUnique(o)

	if err != nil {
		return err
	}

	t.Storage.AddObject(*o)

	if err != nil {
		return err
	}

	return nil
}

func (t *Table) Update(objectM map[string]interface{}) error {
	runMiddleware, update := UpdateMiddleware(t, objectM)

	if runMiddleware {
		return update()
	}

	foundObjectId, err := t.FindExisting(objectM)

	if err != nil {
		return err
	}

	o := t.Storage.GetObject(foundObjectId)

	for _, f := range t.Config.Fields {
		updatedValue, ok := objectM[f.Name]

		if !ok {
			continue
		}

		v, err := idbutil.InterfaceToDBType(updatedValue, f)

		if err != nil {
			return err
		}

		o.M[f.Name] = v
	}

	err = t.allFieldsHaveValues(o)

	if err != nil {
		return err
	}

	err = t.isUnique(o)

	if err != nil {
		return err
	}

	t.Storage.AddObject(*o)

	if err != nil {
		return err
	}

	t.Index.UpdateIndex(*o)

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
		value, err := idbutil.InterfaceToDBType(object[fieldName], f)

		if err != nil {
			return 0, err
		}

		if f.Indexed && f.Unique {
			indexElements := t.Index.Equal(fieldName, value)

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
	gsort.Slice(o, func(i, j int) bool {
		iv := additionalFields[o[i]][fieldName]
		jv := additionalFields[o[j]][fieldName]

		if iv == nil {
			iv = t.Index.GetValue(fieldName, o[i])
		}

		if jv == nil {
			jv = t.Index.GetValue(fieldName, o[j])
		}

		if iv == nil || jv == nil {
			fmt.Println("F")
		}

		return direction == request.ASC && iv.Smaller(jv)
	})

	return o, nil
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
			if len(t.Index.Equal(fieldName, o.M[fieldName])) > 0 {
				return e.FoundExistingObjectWithField(fieldName)
			}
		}
	}

	for _, fieldNames := range t.Config.Options.CombinedUniques {
		var first = true
		var objects object.Objects

		for _, fieldName := range fieldNames {
			if first {
				objects = t.Index.Equal(fieldName, o.M[fieldName])
				first = false
			} else {
				objects = t.and(objects, t.Index.Equal(fieldName, o.M[fieldName]))

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
		if o.M[fieldName] == nil && !f.Null && f.Name != field.InternalObjectIdField {
			return e.ObjectDoesNotHaveValueForField(fieldName)
		}
	}

	return nil
}

func (t *Table) andEqual(andObjects object.Objects, field string, value dbtype.DBType) object.Objects {
	results := object.Objects{}

	for _, andObject := range andObjects {
		if t.Index.GetValue(field, andObject).Equal(value) {
			results = append(results, andObject)
		}
	}

	return results
}

func (t *Table) andNot(andObjects object.Objects, field string, value dbtype.DBType) object.Objects {
	results := object.Objects{}

	for _, andObject := range andObjects {
		if t.Index.GetValue(field, andObject).Not(value) {
			results = append(results, andObject)
		}
	}

	return results
}

func (t *Table) andMatch(andObjects object.Objects, field string, r regexp.Regexp) object.Objects {
	results := object.Objects{}

	for _, andObject := range andObjects {
		if t.Index.GetValue(field, andObject).Matches(r) {
			results = append(results, andObject)
		}
	}

	return results
}

func (t *Table) andSmaller(andObjects object.Objects, field string, value dbtype.DBType) (object.Objects, error) {
	results := object.Objects{}

	for _, andObject := range andObjects {
		if t.Index.GetValue(field, andObject).Smaller(value) {
			results = append(results, andObject)
		}
	}

	return results, nil
}

func (t *Table) andLarger(andObjects object.Objects, field string, value dbtype.DBType) (object.Objects, error) {
	results := object.Objects{}

	for _, andObject := range andObjects {
		if t.Index.GetValue(field, andObject).Larger(value) {
			results = append(results, andObject)
		}
	}

	return results, nil
}

func (t *Table) andBetween(andObjects object.Objects, field string, smaller dbtype.DBType, larger dbtype.DBType) (object.Objects, error) {
	results := object.Objects{}

	for _, andObject := range andObjects {
		if t.Index.GetValue(field, andObject).Between(smaller, larger) {
			results = append(results, andObject)
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

func (t *Table) InterfaceMapToObject(m map[string]interface{}) (*object.Object, error) {
	o := object.Object{
		M: map[string]dbtype.DBType{},
	}

	for _, f := range t.Config.Fields {
		if f.Name == field.InternalObjectIdField {
			continue
		}

		i, ok := m[f.Name]

		if !ok {
			continue
		}

		v, err := idbutil.InterfaceToDBType(i, f)

		if err != nil {
			return nil, err
		}

		o.M[f.Name] = v
	}

	return &o, nil
}

func (t *Table) ObjectToInterfaceMap(o object.Object) (map[string]interface{}, error) {
	m := map[string]interface{}{}

	for _, f := range t.Config.Fields {
		if f.Name == field.InternalObjectIdField {
			continue
		}

		v, ok := o.M[f.Name]

		if !ok {
			continue
		}

		i, err := idbutil.DBTypeToInterface(v, f)

		if err != nil {
			return nil, err
		}

		m[f.Name] = i
	}

	return m, nil
}
