package chef

import (
	"context"
	"fmt"
	"github.com/go-chef/chef"
)

func (s Service) GetEnvironments(ctx context.Context) (interface{}, error) {
	environments, err := s.client.Environments.List()
	if err != nil {
		s.log.Error("failed to list environments")
		return nil, err
	}

	return environments, nil
}

func (s Service) GetEnvironment(ctx context.Context, name string) (*chef.Environment, error) {
	environment, err := s.client.Environments.Get(name)
	// todo: handle 404s as more graceful errors so we can treat 5xx errors differently
	if err != nil {
		s.log.Error(fmt.Sprintf("failed to get environment %s", name))
		return &chef.Environment{}, err
	}

	return environment, nil
}
