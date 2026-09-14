package chef

import (
	"context"
	"testing"

	"github.com/drewhammond/chefbrowser/internal/common/logging"
	"go.uber.org/zap"
)

func newTestMockService() *MockService {
	log := &logging.Logger{Logger: zap.NewNop()}
	return NewMockService(log)
}

func TestMockService_SearchNodesWithDetails(t *testing.T) {
	m := newTestMockService()

	tests := []struct {
		name        string
		query       string
		expectCount bool
	}{
		{
			name:        "match all",
			query:       "*:*",
			expectCount: true,
		},
		{
			name:        "fqdn search",
			query:       "fqdn:*web*",
			expectCount: true,
		},
		{
			name:        "roles search",
			query:       "roles:*base*",
			expectCount: true,
		},
		{
			name:        "addresses search",
			query:       "addresses:*10.*",
			expectCount: true,
		},
		{
			name:        "multi-term OR search",
			query:       "fqdn:*web* OR roles:*database*",
			expectCount: true,
		},
		{
			name:        "no match",
			query:       "fqdn:*nonexistent*",
			expectCount: false,
		},
		{
			name:        "plain text search",
			query:       "web",
			expectCount: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := m.SearchNodesWithDetails(context.Background(), tt.query, 0, 100)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.expectCount && result.Total == 0 {
				t.Errorf("expected results for query %q, got 0", tt.query)
			}
			if !tt.expectCount && result.Total > 0 {
				t.Errorf("expected no results for query %q, got %d", tt.query, result.Total)
			}
		})
	}
}
