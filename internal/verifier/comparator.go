package verifier

import (
	"fmt"
	"strconv"
	"strings"
)

func Compare(targetType, actualRaw, expectedRaw, operator string) (bool, error) {
	actualRaw = strings.TrimSpace(actualRaw)
	switch targetType {
	case "int":
		expectedInt, err := strconv.ParseInt(expectedRaw, 10, 64)
		if err != nil {
			return false, fmt.Errorf("invalid integer in config '%s': %w", expectedRaw, err)
		}

		actualInt, err := strconv.ParseInt(actualRaw, 10, 64)
		if err != nil {
			return false, nil
		}
		return compareOrdered(actualInt, expectedInt, operator)

	case "float":
		expectedFloat, err := strconv.ParseFloat(expectedRaw, 64)
		if err != nil {
			return false, fmt.Errorf("invalid float in config '%s': %w", expectedRaw, err)
		}

		actualFloat, err := strconv.ParseFloat(actualRaw, 64)
		if err != nil {
			return false, nil
		}
		return compareOrdered(actualFloat, expectedFloat, operator)

	case "bool":
		expectedBool, err := strconv.ParseBool(expectedRaw)
		if err != nil {
			return false, fmt.Errorf("invalid boolean in config '%s': %w", expectedRaw, err)
		}

		actualBool, err := strconv.ParseBool(actualRaw)
		if err != nil {
			return false, nil
		}
		return compareBools(actualBool, expectedBool, operator)

	default: // By default try with strings
		return compareStrings(actualRaw, expectedRaw, operator)
	}
}

func compareOrdered[T int64 | float64](actual, expected T, operator string) (bool, error) {
	switch operator {
	case "==", "equals":
		return actual == expected, nil
	case "!=", "notequals":
		return actual != expected, nil
	case ">", "greater":
		return actual > expected, nil
	case ">=", "greaterequals":
		return actual >= expected, nil
	case "<", "less":
		return actual < expected, nil
	case "<=", "lessequals":
		return actual <= expected, nil
	default:
		return false, fmt.Errorf("unknown operator: %s", operator)
	}
}

func compareBools(actual, expected bool, operator string) (bool, error) {
	switch operator {
	case "==", "equals":
		return actual == expected, nil
	case "!=", "notequals":
		return actual != expected, nil
	default:
		return false, fmt.Errorf("operator %s not supported for bool", operator)
	}
}

func compareStrings(actual, expected, operator string) (bool, error) {
	switch operator {
	case "==", "equals":
		return actual == expected, nil
	case "!=", "notequals":
		return actual != expected, nil
	case "contains":
		return strings.Contains(actual, expected), nil
	default:
		return false, fmt.Errorf("operator %s not supported for string", operator)
	}
}
