/*
 * Copyright (c) 2023 Lucas Pape
 */

package functions

import (
	e "github.com/lucasl0st/InfiniteDB/errors"
	"github.com/lucasl0st/InfiniteDB/idblib/object"
	"github.com/lucasl0st/InfiniteDB/idblib/table"
	"github.com/lucasl0st/InfiniteDB/util"
	"math"
)

const fieldNameDistance = "distance"

const radius = 6371

type DistanceFunction struct {
	latitudeFrom  string
	longitudeFrom string

	latitudeToValue  float64
	longitudeToValue float64

	as string
}

func (d *DistanceFunction) Run(
	table *table.Table,
	objects object.Objects,
	additionalFields table.AdditionalFields,
	parameters map[string]interface{},
) (object.Objects, table.AdditionalFields, error) {
	err := d.parseParameters(parameters)

	if err != nil {
		return nil, nil, err
	}

	for _, o := range objects {
		var fromLatitudeValue float64 = 0
		var fromLongitudeValue float64 = 0

		if additionalFields[o][d.latitudeFrom] != nil {
			fromLatitudeValue = additionalFields[o][d.latitudeFrom].(float64)
		} else {
			v, err := util.StringToNumber(table.Index.GetValue(d.latitudeFrom, o))

			if err != nil {
				return nil, nil, err
			}

			fromLatitudeValue = v
		}

		if additionalFields[o][d.longitudeFrom] != nil {
			fromLongitudeValue = additionalFields[o][d.longitudeFrom].(float64)
		} else {
			v, err := util.StringToNumber(table.Index.GetValue(d.longitudeFrom, o))

			if err != nil {
				return nil, nil, err
			}

			fromLongitudeValue = v
		}

		if additionalFields[o] == nil {
			additionalFields[o] = make(map[string]interface{})
		}

		additionalFields[o][d.as] = d.distance(fromLatitudeValue, fromLongitudeValue, d.latitudeToValue, d.longitudeToValue)
	}

	return objects, additionalFields, nil
}

func (d *DistanceFunction) parseParameters(parameters map[string]interface{}) error {
	latFrom, ok := parameters["latitudeFrom"].(string)
	if !ok {
		return e.IsNotAString("latitudeFrom in distance function")
	}

	lonFrom, ok := parameters["longitudeFrom"].(string)
	if !ok {
		return e.IsNotAString("longitudeFrom in distance function")
	}

	latToValue, ok := parameters["latitudeToValue"].(float64)
	if !ok {
		return e.IsNotANumber("latitudeToValue in distance function")
	}

	lonToValue, ok := parameters["longitudeToValue"].(float64)
	if !ok {
		return e.IsNotANumber("longitudeToValue in distance function")
	}

	as, ok := parameters["as"].(string)

	if !ok {
		as = fieldNameDistance
	}

	d.latitudeFrom = latFrom
	d.longitudeFrom = lonFrom
	d.latitudeToValue = latToValue
	d.longitudeToValue = lonToValue
	d.as = as

	return nil
}

func (d *DistanceFunction) distance(fromLatitude float64, fromLongitude float64, toLatitude float64, toLongitude float64) float64 {
	degreesLat := d.degrees2radians(toLatitude - fromLatitude)
	degreesLong := d.degrees2radians(toLongitude - fromLongitude)

	a := math.Sin(degreesLat/2)*math.Sin(degreesLat/2) +
		math.Cos(d.degrees2radians(fromLatitude))*
			math.Cos(d.degrees2radians(toLatitude))*math.Sin(degreesLong/2)*
			math.Sin(degreesLong/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	distance := radius * c

	return distance
}

func (d *DistanceFunction) degrees2radians(degrees float64) float64 {
	return degrees * math.Pi / 180
}
