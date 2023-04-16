/*
 * Copyright (c) 2023 Lucas Pape
 */

package functions

import (
	"encoding/json"
	"github.com/lucasl0st/InfiniteDB/idblib/dbtype"
	"github.com/lucasl0st/InfiniteDB/idblib/object"
	"github.com/lucasl0st/InfiniteDB/idblib/table"
	"github.com/lucasl0st/InfiniteDB/idblib/util"
	"math"
)

const fieldNameDistance = "distance"

const radius = 6371

type DistanceFunction struct {
	latitudeFrom  string
	longitudeFrom string

	latitudeToValue  dbtype.Number
	longitudeToValue dbtype.Number

	as string
}

func (d *DistanceFunction) Run(
	table *table.Table,
	objects object.Objects,
	additionalFields table.AdditionalFields,
	parameters map[string]json.RawMessage,
) (object.Objects, table.AdditionalFields, error) {
	err := d.parseParameters(parameters)

	if err != nil {
		return nil, nil, err
	}

	for _, o := range objects {
		var fromLatitudeValue dbtype.Number
		var fromLongitudeValue dbtype.Number

		if additionalFields[o][d.latitudeFrom] != nil {
			fromLatitudeValue = additionalFields[o][d.latitudeFrom].(dbtype.Number)
		} else {
			fromLatitudeValue = table.Index.GetValue(d.latitudeFrom, o).(dbtype.Number)
		}

		if additionalFields[o][d.longitudeFrom] != nil {
			fromLongitudeValue = additionalFields[o][d.longitudeFrom].(dbtype.Number)
		} else {
			fromLongitudeValue = table.Index.GetValue(d.longitudeFrom, o).(dbtype.Number)
		}

		if additionalFields[o] == nil {
			additionalFields[o] = make(map[string]dbtype.DBType)
		}

		distance, err := d.distance(fromLatitudeValue, fromLongitudeValue, d.latitudeToValue, d.latitudeToValue)

		if err != nil {
			return nil, nil, err
		}

		additionalFields[o][d.as] = distance
	}

	return objects, additionalFields, nil
}

func (d *DistanceFunction) parseParameters(parameters map[string]json.RawMessage) error {
	latFrom, err := util.JsonRawToString(parameters["latitudeFrom"])

	if err != nil {
		return err
	}

	lonFrom, err := util.JsonRawToString(parameters["longitudeFrom"])

	if err != nil {
		return err
	}

	latToValue, err := util.JsonRawToStringNumber(parameters["latitudeToValue"])

	if err != nil {
		return err
	}

	lonToValue, err := util.JsonRawToStringNumber(parameters["latitudeToValue"])

	if err != nil {
		return err
	}

	as := fieldNameDistance

	asP, err := util.JsonRawToString(parameters["as"])

	if asP != nil && len(*asP) > 0 && err == nil {
		as = *asP
	}

	latToValueNumber, err := dbtype.NumberFromString(*latToValue)

	if err != nil {
		return err
	}

	lonToValueNumber, err := dbtype.NumberFromString(*lonToValue)

	d.latitudeFrom = *latFrom
	d.longitudeFrom = *lonFrom
	d.latitudeToValue = latToValueNumber
	d.longitudeToValue = lonToValueNumber
	d.as = as

	return nil
}

func (d *DistanceFunction) distance(fromLatitude dbtype.Number, fromLongitude dbtype.Number, toLatitude dbtype.Number, toLongitude dbtype.Number) (dbtype.Number, error) {
	degreesLat := d.degrees2radians(toLatitude.ToFloat64() - fromLatitude.ToFloat64())
	degreesLong := d.degrees2radians(toLongitude.ToFloat64() - fromLongitude.ToFloat64())

	a := math.Sin(degreesLat/2)*math.Sin(degreesLat/2) +
		math.Cos(d.degrees2radians(fromLatitude.ToFloat64()))*
			math.Cos(d.degrees2radians(toLatitude.ToFloat64()))*math.Sin(degreesLong/2)*
			math.Sin(degreesLong/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	distance := radius * c

	return dbtype.NumberFromFloat64(distance)
}

func (d *DistanceFunction) degrees2radians(degrees float64) float64 {
	return degrees * math.Pi / 180
}
