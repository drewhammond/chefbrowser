package chef

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/drewhammond/chefbrowser/internal/common/logging"
	"github.com/go-chef/chef"
)

type MockService struct {
	log          *logging.Logger
	nodes        []mockNodeData
	roles        map[string]*chef.Role
	environments map[string]*chef.Environment
	cookbooks    map[string][]string
	databags     map[string][]string
	policies     map[string][]string
	policyGroups map[string]map[string]string
	groups       map[string]chef.Group
}

type mockNodeData struct {
	Name        string
	IPAddress   string
	Environment string
	OhaiTime    float64
	RunList     []string
}

func NewMockService(log *logging.Logger) *MockService {
	m := &MockService{log: log}
	m.generateMockData()
	log.Info("Mock service initialized with sample data")
	return m
}

func (m *MockService) generateMockData() {
	m.generateNodes()
	m.generateRoles()
	m.generateEnvironments()
	m.generateCookbooks()
	m.generateDatabags()
	m.generatePolicies()
	m.generateGroups()
}

func (m *MockService) generateNodes() {
	rng := rand.New(rand.NewSource(42))
	now := time.Now()

	prefixes := []string{"web", "app", "db", "cache", "worker", "api", "lb", "monitor", "queue", "search"}
	domains := []string{"prod.example.com", "staging.example.com", "dev.example.com", "qa.example.com"}
	environments := []string{"production", "staging", "development", "qa"}
	roles := []string{"role[base]", "role[webserver]", "role[database]", "role[monitoring]", "role[loadbalancer]"}
	recipes := []string{"recipe[apache2]", "recipe[nginx]", "recipe[mysql]", "recipe[postgresql]", "recipe[redis]", "recipe[nodejs]"}

	m.nodes = make([]mockNodeData, 0, 1500)

	for i := 0; i < 1500; i++ {
		prefix := prefixes[rng.Intn(len(prefixes))]
		domain := domains[rng.Intn(len(domains))]
		env := environments[rng.Intn(len(environments))]

		ipOctet1 := []int{10, 172, 192}[rng.Intn(3)]
		var ipOctet2 int
		if ipOctet1 == 172 {
			ipOctet2 = 16 + rng.Intn(16)
		} else if ipOctet1 == 192 {
			ipOctet2 = 168
		} else {
			ipOctet2 = rng.Intn(256)
		}

		hoursAgo := rng.Intn(168)
		ohaiTime := float64(now.Add(-time.Duration(hoursAgo) * time.Hour).Unix())

		runListSize := 1 + rng.Intn(4)
		runList := make([]string, 0, runListSize)
		runList = append(runList, roles[rng.Intn(len(roles))])
		for j := 1; j < runListSize; j++ {
			runList = append(runList, recipes[rng.Intn(len(recipes))])
		}

		node := mockNodeData{
			Name:        fmt.Sprintf("%s-%03d.%s", prefix, i%100+1, domain),
			IPAddress:   fmt.Sprintf("%d.%d.%d.%d", ipOctet1, ipOctet2, rng.Intn(256), rng.Intn(254)+1),
			Environment: env,
			OhaiTime:    ohaiTime,
			RunList:     runList,
		}
		m.nodes = append(m.nodes, node)
	}

	sort.Slice(m.nodes, func(i, j int) bool {
		return m.nodes[i].Name < m.nodes[j].Name
	})
}

func (m *MockService) generateRoles() {
	m.roles = map[string]*chef.Role{
		"base": {
			Name:        "base",
			Description: "Base role applied to all nodes",
			RunList:     []string{"recipe[base]", "recipe[monitoring::agent]"},
			DefaultAttributes: map[string]interface{}{
				"base": map[string]interface{}{
					"timezone": "UTC",
				},
			},
		},
		"webserver": {
			Name:        "webserver",
			Description: "Web server configuration",
			RunList:     []string{"role[base]", "recipe[nginx]", "recipe[ssl]"},
			DefaultAttributes: map[string]interface{}{
				"nginx": map[string]interface{}{
					"worker_processes": 4,
				},
			},
		},
		"database": {
			Name:        "database",
			Description: "Database server configuration",
			RunList:     []string{"role[base]", "recipe[postgresql]", "recipe[backup]"},
			DefaultAttributes: map[string]interface{}{
				"postgresql": map[string]interface{}{
					"version": "14",
				},
			},
		},
		"monitoring": {
			Name:        "monitoring",
			Description: "Monitoring server with Prometheus and Grafana",
			RunList:     []string{"role[base]", "recipe[prometheus]", "recipe[grafana]"},
		},
		"loadbalancer": {
			Name:        "loadbalancer",
			Description: "Load balancer with HAProxy",
			RunList:     []string{"role[base]", "recipe[haproxy]"},
		},
	}
}

func (m *MockService) generateEnvironments() {
	m.environments = map[string]*chef.Environment{
		"production": {
			Name:        "production",
			Description: "Production environment",
			CookbookVersions: map[string]string{
				"apache2": "= 8.0.0",
				"nginx":   "= 15.0.0",
				"mysql":   ">= 10.0.0",
			},
			DefaultAttributes: map[string]interface{}{
				"environment": "production",
			},
		},
		"staging": {
			Name:        "staging",
			Description: "Staging environment for pre-production testing",
			CookbookVersions: map[string]string{
				"apache2": ">= 8.0.0",
				"nginx":   ">= 15.0.0",
			},
			DefaultAttributes: map[string]interface{}{
				"environment": "staging",
			},
		},
		"development": {
			Name:        "development",
			Description: "Development environment",
			DefaultAttributes: map[string]interface{}{
				"environment": "development",
			},
		},
		"qa": {
			Name:        "qa",
			Description: "QA testing environment",
			DefaultAttributes: map[string]interface{}{
				"environment": "qa",
			},
		},
	}
}

func (m *MockService) generateCookbooks() {
	m.cookbooks = map[string][]string{
		"apache2":    {"8.0.0", "7.5.0", "7.4.0", "7.3.0"},
		"nginx":      {"15.0.0", "14.2.0", "14.1.0"},
		"mysql":      {"10.0.0", "9.5.0", "9.4.0", "9.3.0", "9.2.0"},
		"postgresql": {"11.0.0", "10.5.0", "10.4.0"},
		"redis":      {"8.0.0", "7.5.0", "7.4.0"},
		"nodejs":     {"9.0.0", "8.5.0", "8.4.0"},
		"python":     {"4.0.0", "3.5.0", "3.4.0"},
		"java":       {"12.0.0", "11.5.0", "11.4.0"},
		"monitoring": {"3.0.0", "2.5.0", "2.4.0"},
		"base":       {"5.0.0", "4.5.0", "4.4.0", "4.3.0"},
	}
}

func (m *MockService) generateDatabags() {
	m.databags = map[string][]string{
		"users":       {"admin", "deploy", "backup", "monitoring"},
		"credentials": {"database", "api_keys", "ssl_certs"},
		"app_config":  {"production", "staging", "development"},
	}
}

func (m *MockService) generatePolicies() {
	m.policies = map[string][]string{
		"base":      {"abc123def456", "789ghi012jkl"},
		"webserver": {"mno345pqr678"},
		"database":  {"stu901vwx234", "yz567abc890"},
	}

	m.policyGroups = map[string]map[string]string{
		"production": {
			"base":      "abc123def456",
			"webserver": "mno345pqr678",
		},
		"staging": {
			"base":     "789ghi012jkl",
			"database": "stu901vwx234",
		},
	}
}

func (m *MockService) generateGroups() {
	m.groups = map[string]chef.Group{
		"admins": {
			Name:   "admins",
			Actors: []string{"admin", "superuser"},
			Groups: []string{},
		},
		"users": {
			Name:   "users",
			Actors: []string{"developer1", "developer2", "qa_user"},
			Groups: []string{},
		},
		"clients": {
			Name:   "clients",
			Actors: []string{},
			Groups: []string{},
		},
	}
}

// Node methods

func (m *MockService) GetNodes(ctx context.Context) (*NodeList, error) {
	names := make([]string, len(m.nodes))
	for i, n := range m.nodes {
		names[i] = n.Name
	}
	return &NodeList{Nodes: names}, nil
}

func (m *MockService) SearchNodes(ctx context.Context, q string) (*NodeList, error) {
	var names []string
	searchTerm := strings.ToLower(q)
	for _, n := range m.nodes {
		if strings.Contains(strings.ToLower(n.Name), searchTerm) ||
			strings.Contains(strings.ToLower(n.Environment), searchTerm) ||
			strings.Contains(n.IPAddress, searchTerm) {
			names = append(names, n.Name)
		}
	}
	return &NodeList{Nodes: names}, nil
}

func (m *MockService) GetNodesWithDetails(ctx context.Context, start, pageSize int) (*NodeListResult, error) {
	return m.paginateNodes(m.nodes, start, pageSize)
}

func (m *MockService) SearchNodesWithDetails(ctx context.Context, q string, start, pageSize int) (*NodeListResult, error) {
	var filtered []mockNodeData
	searchTerm := strings.ToLower(q)
	for _, n := range m.nodes {
		if strings.Contains(strings.ToLower(n.Name), searchTerm) ||
			strings.Contains(strings.ToLower(n.Environment), searchTerm) ||
			strings.Contains(n.IPAddress, searchTerm) {
			filtered = append(filtered, n)
		}
	}
	return m.paginateNodes(filtered, start, pageSize)
}

func (m *MockService) paginateNodes(nodes []mockNodeData, start, pageSize int) (*NodeListResult, error) {
	total := len(nodes)
	if start >= total {
		return &NodeListResult{
			Nodes:    []NodeSummary{},
			Total:    total,
			Start:    start,
			PageSize: pageSize,
		}, nil
	}

	end := start + pageSize
	if end > total {
		end = total
	}

	page := nodes[start:end]
	summaries := make([]NodeSummary, len(page))
	for i, n := range page {
		summaries[i] = NodeSummary{
			Name:        n.Name,
			IPAddress:   n.IPAddress,
			Environment: n.Environment,
			OhaiTime:    n.OhaiTime,
		}
	}

	return &NodeListResult{
		Nodes:    summaries,
		Total:    total,
		Start:    start,
		PageSize: pageSize,
	}, nil
}

func (m *MockService) GetNode(ctx context.Context, name string) (*Node, error) {
	for _, n := range m.nodes {
		if n.Name == name {
			node := &Node{
				Node: chef.Node{
					Name:        n.Name,
					Environment: n.Environment,
					RunList:     n.RunList,
					AutomaticAttributes: map[string]interface{}{
						"ipaddress":        n.IPAddress,
						"ohai_time":        n.OhaiTime,
						"fqdn":             n.Name,
						"hostname":         strings.Split(n.Name, ".")[0],
						"platform":         "ubuntu",
						"platform_version": "22.04",
						"lsb": map[string]interface{}{
							"description": "Ubuntu 22.04.3 LTS",
						},
						"chef_packages": map[string]interface{}{
							"chef": map[string]interface{}{
								"version": "18.2.7",
							},
						},
					},
					NormalAttributes:   map[string]interface{}{},
					DefaultAttributes:  map[string]interface{}{},
					OverrideAttributes: map[string]interface{}{},
				},
			}
			node.MergedAttributes = node.MergeAttributes()
			return node, nil
		}
	}
	return nil, fmt.Errorf("node not found: %s", name)
}

// Role methods

func (m *MockService) GetRoles(ctx context.Context) (*RoleList, error) {
	names := make([]string, 0, len(m.roles))
	for name := range m.roles {
		names = append(names, name)
	}
	sort.Strings(names)
	return &RoleList{Roles: names}, nil
}

func (m *MockService) GetRole(ctx context.Context, name string) (*Role, error) {
	if role, ok := m.roles[name]; ok {
		return &Role{Role: role}, nil
	}
	return nil, ErrRoleNotFound
}

// Environment methods

func (m *MockService) GetEnvironments(ctx context.Context) (interface{}, error) {
	result := make(map[string]string)
	for name := range m.environments {
		result[name] = fmt.Sprintf("/environments/%s", name)
	}
	return result, nil
}

func (m *MockService) GetEnvironment(ctx context.Context, name string) (*chef.Environment, error) {
	if env, ok := m.environments[name]; ok {
		return env, nil
	}
	return nil, fmt.Errorf("environment not found: %s", name)
}

// Cookbook methods

func (m *MockService) GetCookbooks(ctx context.Context) (*CookbookListResult, error) {
	var cookbooks []CookbookListItem
	for name, versions := range m.cookbooks {
		cookbooks = append(cookbooks, CookbookListItem{
			Name:     name,
			Versions: versions,
		})
	}
	sort.Slice(cookbooks, func(i, j int) bool {
		return cookbooks[i].Name < cookbooks[j].Name
	})
	return &CookbookListResult{Cookbooks: cookbooks}, nil
}

func (m *MockService) GetLatestCookbooks(ctx context.Context) (*CookbookListResult, error) {
	var cookbooks []CookbookListItem
	for name, versions := range m.cookbooks {
		cookbooks = append(cookbooks, CookbookListItem{
			Name:     name,
			Versions: []string{versions[0]},
		})
	}
	sort.Slice(cookbooks, func(i, j int) bool {
		return cookbooks[i].Name < cookbooks[j].Name
	})
	return &CookbookListResult{Cookbooks: cookbooks}, nil
}

func (m *MockService) GetCookbook(ctx context.Context, name string) (*Cookbook, error) {
	if versions, ok := m.cookbooks[name]; ok {
		return m.GetCookbookVersion(ctx, name, versions[0])
	}
	return nil, ErrCookbookNotFound
}

func (m *MockService) GetCookbookVersion(ctx context.Context, name string, version string) (*Cookbook, error) {
	versions, ok := m.cookbooks[name]
	if !ok {
		return nil, ErrCookbookNotFound
	}

	if version == "_latest" {
		version = versions[0]
	}

	found := false
	for _, v := range versions {
		if v == version {
			found = true
			break
		}
	}
	if !found {
		return nil, ErrCookbookVersionNotFound
	}

	return &Cookbook{
		Cookbook: chef.Cookbook{
			CookbookName: name,
			Name:         fmt.Sprintf("%s-%s", name, version),
			Version:      version,
			Metadata: chef.CookbookMeta{
				Name:        name,
				Version:     version,
				Description: fmt.Sprintf("Mock %s cookbook", name),
				License:     "Apache-2.0",
				Maintainer:  "Mock Maintainer",
			},
			RootFiles: []chef.CookbookItem{
				{Name: "README.md", Path: "README.md"},
				{Name: "metadata.rb", Path: "metadata.rb"},
			},
			Recipes: []chef.CookbookItem{
				{Name: "default.rb", Path: "recipes/default.rb"},
			},
			Attributes: []chef.CookbookItem{
				{Name: "default.rb", Path: "attributes/default.rb"},
			},
		},
	}, nil
}

func (m *MockService) GetCookbookVersions(ctx context.Context, name string) ([]string, error) {
	versions, ok := m.cookbooks[name]
	if !ok {
		return nil, ErrCookbookNotFound
	}
	return versions, nil
}

// Databag methods

func (m *MockService) GetDatabags(ctx context.Context) (interface{}, error) {
	result := make(map[string]string)
	for name := range m.databags {
		result[name] = fmt.Sprintf("/data/%s", name)
	}
	return result, nil
}

func (m *MockService) GetDatabagItems(ctx context.Context, name string) (*chef.DataBagListResult, error) {
	items, ok := m.databags[name]
	if !ok {
		return nil, fmt.Errorf("databag not found: %s", name)
	}

	result := make(chef.DataBagListResult)
	for _, item := range items {
		result[item] = fmt.Sprintf("/data/%s/%s", name, item)
	}
	return &result, nil
}

func (m *MockService) GetDatabagItemContent(ctx context.Context, databag string, item string) (chef.DataBagItem, error) {
	items, ok := m.databags[databag]
	if !ok {
		return nil, fmt.Errorf("databag not found: %s", databag)
	}

	for _, i := range items {
		if i == item {
			return map[string]interface{}{
				"id":      item,
				"databag": databag,
				"data":    "mock_value",
			}, nil
		}
	}
	return nil, fmt.Errorf("databag item not found: %s/%s", databag, item)
}

// Policy methods

func (m *MockService) GetPolicies(ctx context.Context) (chef.PoliciesGetResponse, error) {
	result := make(chef.PoliciesGetResponse)
	for name, revisions := range m.policies {
		revMap := make(map[string]interface{})
		for _, rev := range revisions {
			revMap[rev] = map[string]string{}
		}
		result[name] = chef.Policy{Revisions: revMap}
	}
	return result, nil
}

func (m *MockService) GetPolicy(ctx context.Context, name string) (chef.PolicyGetResponse, error) {
	revisions, ok := m.policies[name]
	if !ok {
		return chef.PolicyGetResponse{}, fmt.Errorf("policy not found: %s", name)
	}

	result := make(chef.PolicyGetResponse)
	for _, rev := range revisions {
		result[rev] = chef.PolicyRevision{}
	}

	return result, nil
}

func (m *MockService) GetPolicyRevision(ctx context.Context, name string, revision string) (chef.RevisionDetailsResponse, error) {
	revisions, ok := m.policies[name]
	if !ok {
		return chef.RevisionDetailsResponse{}, fmt.Errorf("policy not found: %s", name)
	}

	for _, rev := range revisions {
		if rev == revision {
			return chef.RevisionDetailsResponse{
				Name:       name,
				RevisionID: revision,
			}, nil
		}
	}
	return chef.RevisionDetailsResponse{}, fmt.Errorf("policy revision not found: %s/%s", name, revision)
}

func (m *MockService) GetPolicyGroups(ctx context.Context) (chef.PolicyGroupGetResponse, error) {
	result := make(chef.PolicyGroupGetResponse)
	for groupName, policies := range m.policyGroups {
		policyMap := make(map[string]chef.Revision)
		for policyName, revisionID := range policies {
			policyMap[policyName] = chef.Revision{"revision_id": revisionID}
		}
		result[groupName] = chef.PolicyGroup{
			Policies: policyMap,
		}
	}
	return result, nil
}

func (m *MockService) GetPolicyGroup(ctx context.Context, name string) (PolicyGroup, error) {
	policies, ok := m.policyGroups[name]
	if !ok {
		return PolicyGroup{}, fmt.Errorf("policy group not found: %s", name)
	}

	policyMap := make(map[string]chef.Revision)
	for policyName, revisionID := range policies {
		policyMap[policyName] = chef.Revision{"revision_id": revisionID}
	}

	return PolicyGroup{
		PolicyGroup: chef.PolicyGroup{
			Policies: policyMap,
		},
	}, nil
}

// Group methods

func (m *MockService) GetGroups(ctx context.Context) (interface{}, error) {
	result := make(map[string]string)
	for name := range m.groups {
		result[name] = fmt.Sprintf("/groups/%s", name)
	}
	return result, nil
}

func (m *MockService) GetGroup(ctx context.Context, name string) (chef.Group, error) {
	if group, ok := m.groups[name]; ok {
		return group, nil
	}
	return chef.Group{}, fmt.Errorf("group not found: %s", name)
}
