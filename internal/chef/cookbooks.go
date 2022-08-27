package chef

import (
	"context"
	"fmt"
)

func (s Service) GetCookbooks(ctx context.Context) (interface{}, error) {
	cookbooks, err := s.client.Cookbooks.List()
	if err != nil {
		fmt.Println("failed to list cookbooks", err)
		return nil, err
	}
	return cookbooks, nil
}

func (s Service) GetCookbook(ctx context.Context, name string) (interface{}, error) {
	cookbook, err := s.client.Cookbooks.Get(name)
	if err != nil {
		s.log.Error(fmt.Sprintf("failed to get cookbook %s", name))
		return nil, err
	}

	return cookbook, nil
}

func (s Service) GetCookbookVersion(ctx context.Context, name string, version string) (interface{}, error) {
	cookbook, err := s.client.Cookbooks.GetVersion(name, version)
	if err != nil {
		s.log.Error(fmt.Sprintf("failed to get cookbook %s version %s", name, version))
		return nil, err
	}

	return cookbook, nil
}
