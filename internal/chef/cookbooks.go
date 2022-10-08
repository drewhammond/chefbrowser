package chef

import (
	"context"
	"sort"

	"github.com/go-chef/chef"
	"golang.org/x/mod/semver"
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
		return nil, err
	}

	var cookbookList []CookbookListItem

	for j, v := range universe.Books {
		var versions []string

		for q := range v.Versions {
			versions = append(versions, q)
		}

		semver.Sort(versions)
		ReverseSlice(versions)

		cookbook := CookbookListItem{
			Name:     j,
			Versions: versions,
		}

		cookbookList = append(cookbookList, cookbook)
	}

	sort.SliceStable(cookbookList, func(i, j int) bool {
		return cookbookList[i].Name < cookbookList[j].Name
	})

	return &CookbookListResult{Cookbooks: cookbookList}, nil
}

func ReverseSlice[T comparable](s []T) {
	sort.SliceStable(s, func(i, j int) bool {
		return i > j
	})
}

func (s Service) GetLatestCookbooks(ctx context.Context) (*CookbookListResult, error) {
	cookbooks, err := s.client.Cookbooks.List()
	if err != nil {
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
		return nil, err
	}

	return cookbook, nil
}

func (s Service) GetCookbookVersion(ctx context.Context, name string, version string) (*Cookbook, error) {
	cookbook, err := s.client.Cookbooks.GetVersion(name, version)
	if err != nil {
		return nil, err
	}

	return &Cookbook{cookbook}, nil
}
