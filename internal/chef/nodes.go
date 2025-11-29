package chef

import (
	"context"
	"encoding/json"
	"errors"
	"sort"

	"dario.cat/mergo"
	"github.com/go-chef/chef"
)

type NodeList struct {
	Nodes []string `json:"nodes"`
}

type NodeSummary struct {
	Name        string  `json:"name"`
	IPAddress   string  `json:"ipaddress"`
	Environment string  `json:"environment"`
	OhaiTime    float64 `json:"ohai_time"`
}

type NodeListResult struct {
	Nodes    []NodeSummary
	Total    int
	Start    int
	PageSize int
}

type Node struct {
	chef.Node
	MergedAttributes map[string]interface{}
}

var ErrPathNotFound = errors.New("attribute not found at path")

func (s Service) GetNodes(ctx context.Context) (*NodeList, error) {
	nodes, err := s.client.Nodes.List()
	if err != nil {
		return nil, err
	}

	var nl []string

	for i := range nodes {
		nl = append(nl, i)
	}

	sort.Strings(nl)

	return &NodeList{Nodes: nl}, nil
}

func (s Service) SearchNodes(ctx context.Context, q string) (*NodeList, error) {
	partial := map[string]interface{}{
		"name": []string{"name"},
	}
	query, err := s.client.Search.PartialExecJSON("node", q, partial)
	if err != nil {
		return nil, err
	}

	var nodes NodeList

	for _, i := range query.Rows {
		var node Node
		_ = json.Unmarshal(i.Data, &node)
		nodes.Nodes = append(nodes.Nodes, node.Name)
	}

	sort.Strings(nodes.Nodes)

	return &nodes, nil
}

func (s Service) GetNodesWithDetails(ctx context.Context, start, pageSize int) (*NodeListResult, error) {
	return s.searchNodesWithDetails(ctx, "*:*", start, pageSize)
}

func (s Service) SearchNodesWithDetails(ctx context.Context, q string, start, pageSize int) (*NodeListResult, error) {
	return s.searchNodesWithDetails(ctx, q, start, pageSize)
}

func (s Service) searchNodesWithDetails(ctx context.Context, q string, start, pageSize int) (*NodeListResult, error) {
	partial := map[string]interface{}{
		"name":        []string{"name"},
		"environment": []string{"chef_environment"},
		"ipaddress":   []string{"automatic", "ipaddress"},
		"ohai_time":   []string{"automatic", "ohai_time"},
	}

	query := chef.SearchQuery{
		Index: "node",
		Query: q,
		Start: start,
		Rows:  pageSize,
	}

	result, err := query.DoPartialJSON(&s.client, partial)
	if err != nil {
		return nil, err
	}

	nodes := make([]NodeSummary, 0, len(result.Rows))
	for _, row := range result.Rows {
		var data struct {
			Name        string  `json:"name"`
			Environment string  `json:"environment"`
			IPAddress   string  `json:"ipaddress"`
			OhaiTime    float64 `json:"ohai_time"`
		}
		if err := json.Unmarshal(row.Data, &data); err != nil {
			continue
		}
		nodes = append(nodes, NodeSummary{
			Name:        data.Name,
			IPAddress:   data.IPAddress,
			Environment: data.Environment,
			OhaiTime:    data.OhaiTime,
		})
	}

	return &NodeListResult{
		Nodes:    nodes,
		Total:    result.Total,
		Start:    result.Start,
		PageSize: pageSize,
	}, nil
}

func (s Service) GetNode(ctx context.Context, name string) (*Node, error) {
	node, err := s.client.Nodes.Get(name)
	if err != nil {
		return nil, err
	}

	ret := &Node{Node: node}
	ret.MergedAttributes = ret.MergeAttributes()

	return ret, nil
}

// MergeAttributes returns the merged set of all node attributes taking attribute precedence into consideration.
// Ref: https://docs.chef.io/attribute_precedence/
func (s Node) MergeAttributes() map[string]interface{} {
	var attrs map[string]interface{}
	_ = mergo.Merge(&attrs, s.DefaultAttributes, mergo.WithOverride)
	_ = mergo.Merge(&attrs, s.NormalAttributes, mergo.WithOverride)
	_ = mergo.Merge(&attrs, s.OverrideAttributes, mergo.WithOverride)
	_ = mergo.Merge(&attrs, s.AutomaticAttributes, mergo.WithOverride)
	return attrs
}

// GetEffectiveAttributeValue returns the effective attribute value of a given path considering attribute precedence.
func (s Node) GetEffectiveAttributeValue(paths ...string) (interface{}, error) {
	return lookupAttribute(s.MergedAttributes, paths...)
}

// lookupAttribute is a function from go-chef, but we use it differently here since all attributes
// are merged instead of just a single one when requested
func lookupAttribute(attrs map[string]interface{}, paths ...string) (interface{}, error) {
	currentPath, remainingPaths := paths[0], paths[1:]
	if attr, ok := attrs[currentPath]; ok {
		if len(remainingPaths) <= 0 {
			return attr, nil
		}
		return lookupAttribute(attr.(map[string]interface{}), remainingPaths...)
	}

	return nil, ErrPathNotFound
}
