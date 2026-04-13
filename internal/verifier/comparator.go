package verifier

import (
	"fmt"
	"strconv"
	"strings"
)

// Compare сравнивает фактическое значение (из БД) и ожидаемое (из конфига)
func Compare(actual string, expected any, condition string) (bool, error) {
	if condition == "" {
		condition = "eq" // by default check equality
	}

	// Стратегия приведения к строке: всё, что пришло из YAML, превращаем в string
	expectedStr := fmt.Sprintf("%v", expected)

	switch condition {
	case "eq":
		return actual == expectedStr, nil
	case "not_empty":
		return actual != "" && actual != "0" && actual != "[]" && actual != "false" && actual != "null", nil
	case "gt", "lt":
		// Пытаемся безопасно распарсить обе строки как float64
		actualFloat, err1 := strconv.ParseFloat(actual, 64)
		expectedFloat, err2 := strconv.ParseFloat(expectedStr, 64)
		if err1 != nil || err2 != nil {
			return false, fmt.Errorf("невозможно сравнить как числа: actual='%s', expected='%s'", actual, expectedStr)
		}

		if condition == "gt" {
			return actualFloat > expectedFloat, nil
		}
		return actualFloat < expectedFloat, nil
	case "contains":
		return strings.Contains(actual, expectedStr), nil
	default:
		return false, fmt.Errorf("неизвестный оператор сравнения: %s", condition)
	}
}
