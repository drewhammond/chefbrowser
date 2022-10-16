package util

import (
	"reflect"
	"testing"
)

func TestMakeJSONPath(t *testing.T) {

	input := map[string]interface{}{
		"stringkey": "stringval",
		"somearr":   []string{"one", "two", "three"},
		"nested": map[string]interface{}{
			"foo":   "bar",
			"hello": "world",
			"bool":  false,
			"int":   123,
			"deep": map[string]interface{}{
				"nest": "value",
			},
		},
	}

	expected := map[string]interface{}{
		"$.stringkey":        "stringval",
		"$.somearr":          []string{"one", "two", "three"},
		"$.nested.foo":       "bar",
		"$.nested.bool":      false,
		"$.nested.int":       123,
		"$.nested.hello":     "world",
		"$.nested.deep.nest": "value",
	}

	actual := MakeJSONPath(input, "$")

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("failed, in: %v, expected: %v, actual: %v", input, expected, actual)
	}
}
