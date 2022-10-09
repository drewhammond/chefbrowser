package chef

import (
	"context"
	"fmt"
	"sort"

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
	query, err := s.client.Search.NewQuery("node", q)
	if err != nil {
		return nil, err
	}

	fmt.Println(query)
	panic("finish me")
}

func (s Service) GetNode(ctx context.Context, name string) (*Node, error) {
	node, err := s.client.Nodes.Get(name)
	if err != nil {
		return nil, err
	}

	return &Node{node}, nil
}
