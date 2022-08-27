package chef

import (
	"context"
)

func (s Service) GetGroups(ctx context.Context) (interface{}, error) {
	groups, err := s.client.Groups.List()
	if err != nil {
		s.log.Error("failed to list groups")
		return nil, err
	}

	return groups, nil
}

func (s Service) GetGroup(ctx context.Context, name string) (interface{}, error) {
	group, err := s.client.Groups.Get(name)
	if err != nil {
		s.log.Error("failed to list group")
		return nil, err
	}

	return group, nil
}
