package chef

import "testing"

func Test_normalizeChefURL(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		input    string
		expected string
	}{
		{
			"http://localhost/organizations/foo",
			"http://localhost/organizations/foo/",
		},
		{
			"http://localhost/foo",
			"http://localhost/foo",
		},
		{
			"http://localhost/organizations/foo/",
			"http://localhost/organizations/foo/",
		},
	}
	for _, tt := range tests {
		t.Run("foo", func(t *testing.T) {
			if got := normalizeChefURL(tt.input); got != tt.expected {
				t.Errorf("normalizeChefURL() = %v, want %v", got, tt.expected)
			}
		})
	}
}
