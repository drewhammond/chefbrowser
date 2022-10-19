package chef

import (
	"errors"
	"testing"

	"github.com/go-chef/chef"
)

func TestGetEffectiveAttributes(t *testing.T) {
	tests := []struct {
		name     string
		node     Node
		paths    []string
		expected interface{}
		err      error
	}{
		{
			"simple top level",
			Node{
				Node: chef.Node{
					NormalAttributes:  map[string]interface{}{"foo": "normal"},
					DefaultAttributes: map[string]interface{}{"foo": "default"},
				},
			},
			[]string{"foo"},
			"normal",
			nil,
		},
		{
			"override wins over default",
			Node{
				Node: chef.Node{
					DefaultAttributes:  map[string]interface{}{"foo": "default"},
					OverrideAttributes: map[string]interface{}{"foo": "override"},
				},
			},
			[]string{"foo"},
			"override",
			nil,
		},
		{
			"deep merge",
			Node{
				Node: chef.Node{
					NormalAttributes: map[string]interface{}{
						"foo": map[string]interface{}{
							"bar": "overwritten",
						},
					},
					DefaultAttributes: map[string]interface{}{
						"foo": map[string]interface{}{
							"bar": "original",
						},
					},
				},
			},
			[]string{"foo", "bar"},
			"overwritten",
			nil,
		},
		{
			"mixed data types",
			Node{
				Node: chef.Node{
					NormalAttributes: map[string]interface{}{
						"foo": map[string]interface{}{
							"bar": true,
						},
					},
					DefaultAttributes: map[string]interface{}{
						"foo": map[string]interface{}{
							"bar": "original",
						},
					},
				},
			},
			[]string{"foo", "bar"},
			true,
			nil,
		},
		{
			"multi merge",
			Node{
				Node: chef.Node{
					NormalAttributes: map[string]interface{}{
						"foo": map[string]interface{}{
							"bar": "normal",
						},
					},
					OverrideAttributes: map[string]interface{}{
						"foo": map[string]interface{}{
							"bar": "override",
						},
					},
					AutomaticAttributes: map[string]interface{}{
						"foo": map[string]interface{}{
							"bar": "automatic",
						},
					},
					DefaultAttributes: map[string]interface{}{
						"foo": map[string]interface{}{
							"bar": "default",
						},
					},
				},
			},
			[]string{"foo", "bar"},
			"automatic",
			nil,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			tt.node.MergedAttributes = tt.node.MergeAttributes()
			output, err := tt.node.GetEffectiveAttributeValue(tt.paths...)
			if err != nil {
				if tt.err == nil {
					t.Errorf("unxpected error, expected: %v, actual: %v", tt.err, err)
				}

				if !errors.Is(err, tt.err) {
					t.Errorf("unexpected error type, expected: %v, actual: %v", tt.err, err)
				}
			}
			if tt.err != nil && err == nil {
				t.Errorf("should have error, expected: %v, actual: %v", tt.err, err)
			}

			if output != tt.expected {
				t.Errorf("unexpected result, paths: %v, expected: %v, actual: %v", tt.paths, tt.expected, output)
			}
		})
	}
}
