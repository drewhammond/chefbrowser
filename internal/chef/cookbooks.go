package chef

import (
	"context"
	"fmt"
	"github.com/go-chef/chef"
)

type CookbookVersion struct {
	Version string
}

//type Cookbook struct {
//	chef.CookbookMeta
//	Name       string
//	Metadata   string // todo
//	Recipes    string // todo
//	Attributes string // todo
//	Templates  string // todo
//	Resources  string // todo
//	Versions   []CookbookVersion
//}

type CookbookListItem struct {
	Name         string   `json:"name"`
	Versions     []string `json:"versions"`
	Dependencies []string `json:"dependencies,omitempty"`
}

type CookbookListResult struct {
	Cookbooks []CookbookListItem `json:"cookbooks"`
}

type Cookbook struct {
	chef.Cookbook
}

func (s Service) GetCookbooks(ctx context.Context) (*CookbookListResult, error) {
	universe, err := s.client.Universe.Get()
	if err != nil {
		fmt.Println("failed to list cookbooks", err)
		return nil, err
	}

	var cookbookList []CookbookListItem

	for j, v := range universe.Books {
		var versions []string

		for q, _ := range v.Versions {
			versions = append(versions, q)
		}

		cookbook := CookbookListItem{
			Name:     j,
			Versions: versions,
		}

		cookbookList = append(cookbookList, cookbook)
	}

	return &CookbookListResult{Cookbooks: cookbookList}, nil
}

func (s Service) GetLatestCookbooks(ctx context.Context) (*CookbookListResult, error) {

	cookbooks, err := s.client.Cookbooks.List()
	if err != nil {
		fmt.Println("failed to list cookbooks", err)
		return nil, err
	}

	var cookbookList []CookbookListItem

	for j, v := range cookbooks {
		var versions []string
		for _, k := range v.Versions {
			versions = append(versions, k.Version)
		}

		cookbook := CookbookListItem{
			Name:     j,
			Versions: versions,
		}

		cookbookList = append(cookbookList, cookbook)
	}

	return &CookbookListResult{Cookbooks: cookbookList}, nil
}

// GetCookbook should get the latest version of the cookbook
func (s Service) GetCookbook(ctx context.Context, name string) (*Cookbook, error) {
	cookbook, err := s.GetCookbookVersion(ctx, name, "_latest")
	if err != nil {
		s.log.Error(fmt.Sprintf("failed to get latest version of cookbook %s", name))
		return nil, err
	}

	return cookbook, nil
}

func (s Service) GetCookbookVersion(ctx context.Context, name string, version string) (*Cookbook, error) {
	cookbook, err := s.client.Cookbooks.GetVersion(name, version)
	if err != nil {
		s.log.Error(fmt.Sprintf("failed to get cookbook %s version %s", name, version))
		return nil, err
	}

	return &Cookbook{cookbook}, nil
}
