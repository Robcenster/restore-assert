package verifier

import (
	"fmt"
	"strconv"
	"strings"
)

// Compare compares the actual value (from the database) with the expected value (from the configuration)
func Compare(actual string, expected any, condition string) (bool, error) {
	if condition == "" {
		condition = "eq" // by default check equality
	}

	expectedStr := fmt.Sprintf("%v", expected)

	switch condition {
	case "eq":
		return actual == expectedStr, nil
	case "not_empty":
		return actual != "" && actual != "0" && actual != "[]" && actual != "false" && actual != "null", nil
	case "gt", "lt":
		actualFloat, err1 := strconv.ParseFloat(actual, 64)
		expectedFloat, err2 := strconv.ParseFloat(expectedStr, 64)
		if err1 != nil || err2 != nil {
			return false, fmt.Errorf("cannot be compared as numbers: actual='%s', expected='%s'", actual, expectedStr)
		}

		if condition == "gt" {
			return actualFloat > expectedFloat, nil
		}
		return actualFloat < expectedFloat, nil
	case "contains":
		return strings.Contains(actual, expectedStr), nil
	default:
		return false, fmt.Errorf("unknown comparison operator: %s", condition)
	}
}
