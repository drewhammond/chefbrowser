package chef

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"dario.cat/mergo"
	"github.com/go-chef/chef"
)

type NodeList struct {
	Nodes []string `json:"nodes"`
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

// SearchNodes runs a Chef search against the node index. A query without a
// "field:value" pair is expanded to the same fuzzy match knife uses.
// A limit > 0 asks the server for at most that many rows in a single request
// instead of paging through every match.
func (s Service) SearchNodes(ctx context.Context, q string, limit int) (*NodeList, error) {
	if !strings.Contains(q, ":") {
		q = fuzzifySearchStr(q)
	}
	partial := map[string]interface{}{
		"name": []string{"name"},
	}
	var query chef.JSearchResult
	var err error
	if limit > 0 {
		sq := chef.SearchQuery{Index: "node", Query: q, SortBy: "X_CHEF_id_CHEF_X asc", Rows: limit}
		query, err = sq.DoPartialJSON(&s.client, partial)
	} else {
		query, err = s.client.Search.PartialExecJSON("node", q, partial)
	}
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

// fuzzifySearchStr mimics the fuzzy search functionality
// provided by chef https://github.com/chef/chef/blob/main/lib/chef/search/query.rb#L109
func fuzzifySearchStr(s string) string {
	format := []string{
		"tags:*%v*",
		"roles:*%v*",
		"fqdn:*%v*",
		"addresses:*%v*",
		"policy_name:*%v*",
		"policy_group:*%v*",
	}
	var b strings.Builder
	for i, f := range format {
		if i > 0 {
			b.WriteString(" OR ")
		}
		b.WriteString(fmt.Sprintf(f, s))
	}
	return b.String()
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
