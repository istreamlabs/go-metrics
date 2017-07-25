package metrics

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strings"
)

// Combine two maps, with the second one overriding duplicate values.
func combine(original, override map[string]string) map[string]string {
	// We know the size must be at least the length of the existing tag map, but
	// since values can be overridden we cannot assume the length is the sum of
	// both inputs.
	combined := make(map[string]string, len(original))

	for k, v := range original {
		combined[k] = v
	}
	for k, v := range override {
		combined[k] = v
	}

	return combined
}

// Converts a map to an array of strings like `key:value`.
func mapToStrings(tagMap map[string]string) []string {
	tags := make([]string, 0, len(tagMap))

	for k, v := range tagMap {
		tags = append(tags, fmt.Sprintf("%s:%s", k, v))
	}

	return tags
}

// convertType converts a value into an specific type if possible, otherwise
// panics. The returned interface is guaranteed to cast properly.
func convertType(value interface{}, toType reflect.Type) interface{} {
	v := reflect.Indirect(reflect.ValueOf(value))
	if !v.Type().ConvertibleTo(toType) {
		panic(fmt.Sprintf("cannot convert %v to %v", v.Type(), toType))
	}
	return v.Convert(toType).Interface()
}

// toFloat64 converts a value into a float64 if possible, otherwise panics.
func toFloat64(value interface{}) float64 {
	return convertType(value, reflect.TypeOf(float64(0.0))).(float64)
}

// getBlurb returns a line of text from the given file and line number. Useful
// for additional context in stack traces.
func getBlurb(fname string, lineno int) string {
	file, err := os.Open(fname)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	current := 1
	var blurb string
	for scanner.Scan() {
		if current == lineno {
			blurb = strings.Trim(scanner.Text(), " \t")
			break
		}
		current++
	}
	return blurb
}
