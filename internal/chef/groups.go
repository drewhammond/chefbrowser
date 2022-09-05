package chef

import (
	"context"

	"github.com/go-chef/chef"
)

func (s Service) GetGroups(ctx context.Context) (interface{}, error) {
	groups, err := s.client.Groups.List()
	if err != nil {
		s.log.Error("failed to list groups")
		return groups, err
	}

	return groups, nil
}

func (s Service) GetGroup(ctx context.Context, name string) (chef.Group, error) {
	group, err := s.client.Groups.Get(name)
	if err != nil {
		s.log.Error("failed to list group")
		return group, err
	}

	return group, nil
}
