package utils

import (
	"strconv"
)

// ParseFloat parses a string to a float64
func ParseFloat(value string) float64 {
	parsedValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0.0
	}
	return parsedValue
}
