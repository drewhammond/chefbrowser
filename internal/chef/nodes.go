package chef

import (
	"context"
	"encoding/json"
	"sort"

	"github.com/drewhammond/chefbrowser/internal/util"
	"github.com/go-chef/chef"
)

type NodeList struct {
	Nodes []string `json:"nodes"`
}

type Node struct {
	chef.Node
}

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

func (s Service) GetNode(ctx context.Context, name string) (*Node, error) {
	node, err := s.client.Nodes.Get(name)
	if err != nil {
		return nil, err
	}

	node.AutomaticAttributes = util.MakeJSONPath(node.AutomaticAttributes, "$")
	node.NormalAttributes = util.MakeJSONPath(node.NormalAttributes, "$")
	node.DefaultAttributes = util.MakeJSONPath(node.DefaultAttributes, "$")
	node.OverrideAttributes = util.MakeJSONPath(node.OverrideAttributes, "$")

	return &Node{node}, nil
}
