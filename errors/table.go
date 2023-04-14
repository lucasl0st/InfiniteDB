/*
 * Copyright (c) 2023 Lucas Pape
 */

package errors

import (
	"errors"
	"fmt"
	"github.com/lucasl0st/InfiniteDB/request"
)

func CannotFindField(fieldName string) error {
	return errors.New(fmt.Sprintf("cannot find field \"%s\" in table", fieldName))
}

func FieldHasUnsupportedTypeForThisFunction(fieldName string) error {
	return errors.New(fmt.Sprintf("the field \"%s\" has an unsupported type for this function", fieldName))
}

func NotEnoughValuesForOperator(operator request.Operator) error {
	return errors.New(fmt.Sprintf("not enough values for operator %v", operator))
}

func NotAValidOperator() error {
	return errors.New("not a valid operator")
}

func CouldNotFindObjectWithAtLeastOneIndexedAndUniqueValue() error {
	return errors.New("could not find object with at least one indexed and unique value")
}

func FoundExistingObjectWithField(fieldName string) error {
	return errors.New(fmt.Sprintf("found existing object with field %s", fieldName))
}

func FoundExistingObjectWithCombinedUniques() error {
	return errors.New("found existing object with combined uniques")
}

func ObjectDoesNotHaveValueForField(fieldName string) error {
	return errors.New(fmt.Sprintf("object does not have value for field %s and field cannot be null", fieldName))
}

func ValueForOperatorMustBeString(operator request.Operator) error {
	return errors.New(fmt.Sprintf("value must be string for operator %s", operator))
}
