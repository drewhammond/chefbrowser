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

func (s Service) GetPolicyGroups(ctx context.Context) (chef.PolicyGroupGetResponse, error) {
	policyGroups, err := s.client.PolicyGroups.List()
	if err != nil {
		return policyGroups, err
	}

	return policyGroups, nil
}

func (s Service) GetPolicyGroup(ctx context.Context, name string) (chef.PolicyGroup, error) {
	policyGroup, err := s.client.PolicyGroups.Get(name)
	if err != nil {
		return policyGroup, err
	}

	return policyGroup, nil
}
