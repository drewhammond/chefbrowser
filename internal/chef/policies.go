package chef

import (
	"context"

	"github.com/go-chef/chef"
)

type PolicyGroup struct {
	chef.PolicyGroup
}

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

func (s Service) GetPolicyRevision(ctx context.Context, name string, revision string) (chef.RevisionDetailsResponse, error) {
	policyRevision, err := s.client.Policies.GetRevisionDetails(name, revision)
	if err != nil {
		return policyRevision, err
	}

	return policyRevision, nil
}

func (s Service) GetPolicyGroups(ctx context.Context) (chef.PolicyGroupGetResponse, error) {
	policyGroups, err := s.client.PolicyGroups.List()
	if err != nil {
		return policyGroups, err
	}

	return policyGroups, nil
}

func (s Service) GetPolicyGroup(ctx context.Context, name string) (PolicyGroup, error) {
	policyGroup, err := s.client.PolicyGroups.Get(name)
	resp := PolicyGroup{policyGroup}
	if err != nil {
		return resp, err
	}

	return resp, nil
}
