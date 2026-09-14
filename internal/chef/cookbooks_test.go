package chef

import (
	"reflect"
	"testing"
)

func TestReverseSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "basic reverse",
			input:    []string{"1.0.0", "2.0.0", "3.0.0"},
			expected: []string{"3.0.0", "2.0.0", "1.0.0"},
		},
		{
			name:     "single element",
			input:    []string{"1.0.0"},
			expected: []string{"1.0.0"},
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "two elements",
			input:    []string{"a", "b"},
			expected: []string{"b", "a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := make([]string, len(tt.input))
			copy(input, tt.input)
			ReverseSlice(input)
			if !reflect.DeepEqual(input, tt.expected) {
				t.Errorf("ReverseSlice() = %v, want %v", input, tt.expected)
			}
		})
	}
}
