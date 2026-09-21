package chef

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/go-chef/chef"
)

// newTestService returns a Service whose chef client points at srv.
func newTestService(t *testing.T, srv *httptest.Server) Service {
	t.Helper()
	k, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	key := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)})
	client, err := chef.NewClient(&chef.Config{Name: "test", Key: string(key), BaseURL: srv.URL + "/"})
	if err != nil {
		t.Fatal(err)
	}
	return Service{client: *client}
}

func TestSearchNodes(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		limit    int
		wantRows string // rows= sent to the server
		wantQ    string // substring expected in q= sent to the server
	}{
		{"limit sends rows and skips paging", "cookbooks:nginx", 1, "1", "cookbooks:nginx"},
		{"no limit uses default page size", "cookbooks:nginx", 0, "1000", "cookbooks:nginx"},
		{"bare term is fuzzified", "web", 0, "1000", "fqdn:*web*"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var calls int
			var gotRows, gotQ string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				gotRows = r.URL.Query().Get("rows")
				gotQ = r.URL.Query().Get("q")
				fmt.Fprint(w, `{"total":2,"start":0,"rows":[{"url":"","data":{"name":"node-b"}},{"url":"","data":{"name":"node-a"}}]}`)
			}))
			defer srv.Close()

			s := newTestService(t, srv)
			got, err := s.SearchNodes(t.Context(), tt.query, tt.limit)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if calls != 1 {
				t.Errorf("expected exactly 1 request to the server, got %d", calls)
			}
			if gotRows != tt.wantRows {
				t.Errorf("rows param: expected %q, got %q", tt.wantRows, gotRows)
			}
			if !strings.Contains(gotQ, tt.wantQ) {
				t.Errorf("q param: expected to contain %q, got %q", tt.wantQ, gotQ)
			}
			if want := []string{"node-a", "node-b"}; !reflect.DeepEqual(got.Nodes, want) {
				t.Errorf("nodes: expected %v (sorted), got %v", want, got.Nodes)
			}
		})
	}
}

func TestFuzzifySearchStr(t *testing.T) {
	got := fuzzifySearchStr("web")
	want := "tags:*web* OR roles:*web* OR fqdn:*web* OR addresses:*web* OR policy_name:*web* OR policy_group:*web*"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

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
