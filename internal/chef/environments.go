package chef

import (
	"context"
	"errors"

	"github.com/go-chef/chef"
)

var ErrEnvironmentNotFound = errors.New("environment not found")

func (s Service) GetEnvironments(ctx context.Context) (interface{}, error) {
	environments, err := s.client.Environments.List()
	if err != nil {
		return nil, err
	}

	return environments, nil
}

func (s Service) GetEnvironment(ctx context.Context, name string) (*chef.Environment, error) {
	environment, err := s.client.Environments.Get(name)
	// todo: handle 404s as more graceful errors so we can treat 5xx errors differently
	if err != nil {
		return &chef.Environment{}, ErrEnvironmentNotFound
	}

	return environment, nil
}
