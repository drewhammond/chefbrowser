package chef

import (
	"context"

	"github.com/go-chef/chef"
)

func (s Service) GetPolicies(ctx context.Context) (chef.PoliciesGetResponse, error) {
	policies, err := s.client.Policies.List()
	if err != nil {
		return policies, err
	}

	return policies, nil
}

func (s Service) GetPolicy(ctx context.Context, name string) (chef.PolicyGetResponse, error) {
	policy, err := s.client.Policies.Get(name)
	if err != nil {
		return policy, err
	}

	return policy, nil
}
