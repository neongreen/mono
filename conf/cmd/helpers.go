package main

import "strconv"

// parseValue attempts to parse a string value into the appropriate type
func parseValue(value string) interface{} {

	if value == "true" || value == "false" {
		return value == "true"
	}

	if intVal, err := strconv.Atoi(value); err == nil {
		return intVal
	}

	if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
		return floatVal
	}

	return value
}
