package chef

import (
	"context"
	"fmt"
)

func (s Service) GetNodes(ctx context.Context) (interface{}, error) {
	nodes, err := s.client.Nodes.List()
	if err != nil {
		fmt.Println("failed to list nodes", err)
		return nil, err
	}
	return nodes, nil
}

func (s Service) GetNode(ctx context.Context, name string) (interface{}, error) {
	node, err := s.client.Nodes.Get(name)
	if err != nil {
		s.log.Error(fmt.Sprintf("failed to get node %s", name))
		return nil, err
	}

	return node, nil
}
