package chef

import (
	"context"
	"errors"
	"sort"

	"github.com/go-chef/chef"
)

var ErrRoleNotFound = errors.New("role not found")

type RoleList struct {
	Roles []string `json:"roles"`
}

type Role struct {
	*chef.Role
}

// GetRole will return a single named role
func (s Service) GetRole(ctx context.Context, name string) (*Role, error) {
	role, err := s.client.Roles.Get(name)
	if err != nil {
		return nil, ErrRoleNotFound
	}

	return &Role{role}, nil
}

// GetRoles will return a list of all roles found on the server
func (s Service) GetRoles(ctx context.Context) (*RoleList, error) {
	roles, err := s.client.Roles.List()
	if err != nil {
		return nil, err
	}

	var rl []string
	for i := range *roles {
		rl = append(rl, i)
	}
	sort.Strings(rl)

	return &RoleList{rl}, nil
}
