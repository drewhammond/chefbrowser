package chef

import (
	"context"
	"fmt"
)

// GetRole will return a single named role
func (s Service) GetRole(ctx context.Context, name string) (interface{}, error) {
	role, err := s.client.Roles.Get(name)
	if err != nil {
		s.log.Error(fmt.Sprintf("failed to get role %s", name))
		return nil, err
	}

	return role, nil
}

// GetRoles will return a list of all roles found on the server
func (s Service) GetRoles(ctx context.Context) (interface{}, error) {
	roles, err := s.client.Roles.List()
	if err != nil {
		fmt.Println("failed to list roles", err)
		return nil, err
	}
	return roles, nil
}
