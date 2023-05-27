/*
 * Copyright (c) 2023 Lucas Pape
 */

package covid19

type Data struct {
	State     string
	County    string
	AgeGroup  string
	Gender    string
	Date      string
	Cases     int64
	Deaths    int64
	Recovered int64
}
