package domain

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

func EnumString[T ~int](val T, stringsMap []string) string {
	idx := int(val)
	if idx < 0 || idx >= len(stringsMap) {
		return stringUnknown
	}

	return stringsMap[idx]
}

func EnumIsValid[T ~int](val, maxVal T) bool {
	return val >= 0 && val <= maxVal
}

func EnumValues[T ~int](maxVal T) []T {
	maxInt := int(maxVal)

	values := make([]T, maxInt+1)
	for i := 0; i <= maxInt; i++ {
		values[i] = T(i)
	}

	return values
}

func EnumMarshalJSON[T ~int](val T, stringsMap []string) ([]byte, error) {
	return json.Marshal(EnumString(val, stringsMap))
}

func EnumUnmarshalJSON[T ~int](
	data []byte,
	target *T,
	stringsMap []string,
	name string,
) error {
	var s string

	err := json.Unmarshal(data, &s)
	if err != nil {
		return fmt.Errorf("%s must be a string: %w", name, err)
	}

	upper := strings.ToUpper(s)
	for i, str := range stringsMap {
		if strings.ToUpper(str) == upper {
			*target = T(i)

			return nil
		}
	}

	return fmt.Errorf("invalid %s: %s. Valid: %v", name, s, stringsMap)
}

func EnumMarshalYAML[T ~int](val T, stringsMap []string) (any, error) {
	return EnumString(val, stringsMap), nil
}

func EnumUnmarshalYAML[T ~int](
	value *yaml.Node,
	target *T,
	stringsMap []string,
	name string,
) error {
	var s string

	if err := value.Decode(&s); err == nil {
		if i, atoiErr := strconv.Atoi(s); atoiErr == nil {
			if i >= 0 && i < len(stringsMap) {
				*target = T(i)

				return nil
			}

			return fmt.Errorf("invalid %s: %d. Valid: 0-%d (%v)", name, i, len(stringsMap)-1, stringsMap)
		}

		upper := strings.ToUpper(s)
		for i, str := range stringsMap {
			if strings.ToUpper(str) == upper {
				*target = T(i)

				return nil
			}
		}

		return fmt.Errorf("invalid %s: %s. Valid: %v", name, s, stringsMap)
	}

	var i int
	if err := value.Decode(&i); err == nil {
		if i >= 0 && i < len(stringsMap) {
			*target = T(i)

			return nil
		}

		return fmt.Errorf("invalid %s: %d. Valid: 0-%d (%v)", name, i, len(stringsMap)-1, stringsMap)
	}

	return fmt.Errorf("cannot parse %s: expected string or int", name)
}
