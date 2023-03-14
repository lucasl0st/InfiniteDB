/*
 * Copyright (c) 2023 Lucas Pape
 */

package util

import (
	"math"
	"testing"
)

func TestRandomBoolean(t *testing.T) {
	c := 10000

	numberOfTrue := 0
	numberOfFalse := 0

	for i := 0; i < c; i++ {
		if RandomBoolean() {
			numberOfTrue++
		} else {
			numberOfFalse++
		}
	}

	if math.Abs(float64(numberOfTrue-numberOfFalse)) > float64(c/10) {
		t.Errorf("number of random trues not in relation to number of falses, trues: %v falses: %v", numberOfTrue, numberOfFalse)
	}
}

func TestRandomString(t *testing.T) {
	c := 10000

	doubles := 0
	var randomStrings []string

	for i := 0; i < c; i++ {
		randomStrings = append(randomStrings, RandomString(32))
	}

	for _, s := range randomStrings {
		found := false

		for _, s2 := range randomStrings {
			if s == s2 {
				if found {
					doubles++
				} else {
					found = true
				}
			}
		}
	}

	if doubles > c/1000 {
		t.Errorf("number of double random strings exceeds maximum")
	}
}

func TestRandomFloat(t *testing.T) {
	c := 10000

	doubles := 0
	var randomFloats []float64

	for i := 0; i < c; i++ {
		randomFloats = append(randomFloats, RandomFloat())
	}

	for _, s := range randomFloats {
		found := false

		for _, s2 := range randomFloats {
			if s == s2 {
				if found {
					doubles++
				} else {
					found = true
				}
			}
		}
	}

	if doubles > c/1000 {
		t.Errorf("number of double random floats exceeds maximum")
	}
}
