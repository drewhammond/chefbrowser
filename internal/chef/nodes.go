package chef

import (
	"context"
	"fmt"
	"sort"

	"github.com/go-chef/chef"
	"go.uber.org/zap"
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
		s.log.Error("failed to list nodes", zap.Error(err))
		return nil, err
	}

	var nl []string

	for i := range nodes {
		nl = append(nl, i)
	}

	sort.Strings(nl)

	return &NodeList{Nodes: nl}, nil
}

func (s Service) GetNode(ctx context.Context, name string) (*Node, error) {
	node, err := s.client.Nodes.Get(name)
	if err != nil {
		s.log.Error(fmt.Sprintf("failed to get node %s", name))
		return nil, err
	}

	return &Node{node}, nil
}
