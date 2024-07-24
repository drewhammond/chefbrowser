package chef

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/go-chef/chef"
	"golang.org/x/mod/semver"
)

var (
	ErrCookbookNotFound        = errors.New("cookbook not found")
	ErrCookbookVersionNotFound = errors.New("cookbook version not found")
	ErrCookbookFileNotFound    = errors.New("cookbook file not found")
	ErrInternalServerError     = errors.New("internal server error")
)

type CookbookVersion struct {
	Version string
}

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

type CookbookMeta struct {
	chef.CookbookMeta
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
			// semver.Sort requires versions to be prefixed with "v"
			versions = append(versions, "v"+q)
		}

		semver.Sort(versions)

		// strip the leading "v" now that we're properly sorted
		for i, _ := range versions {
			versions[i] = versions[i][1:]
		}

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
		if cerr, ok := err.(*chef.ErrorResponse); ok {
			if cerr.StatusCode() == 404 {
				return nil, ErrCookbookVersionNotFound
			}
		}
		s.log.Error(err.Error())
		return nil, ErrInternalServerError
	}

	return &Cookbook{cookbook}, nil
}

func (s Cookbook) GetFile(ctx context.Context, client *http.Client, path string) (string, error) {
	t := strings.SplitN(path, "/", 2)[0]
	var loc []chef.CookbookItem
	switch t {
	case "attributes":
		loc = s.Attributes
	case "recipes":
		loc = s.Recipes
	case "resources":
		loc = s.Resources
	case "libraries":
		loc = s.Libraries
	case "providers":
		loc = s.Providers
	case "templates":
		loc = s.Templates
	case "files":
		loc = s.Files
	default:
		loc = s.RootFiles
	}
	for _, f := range loc {
		if f.Path == path {
			content, err := downloadFile(client, f.Url)
			if err != nil {
				return "", err
			}
			return string(content), err
		}
	}

	return "", ErrCookbookFileNotFound
}

func downloadFile(client *http.Client, url string) ([]byte, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return body, nil
}

func (s Cookbook) GetReadme(ctx context.Context, client *http.Client) (string, error) {
	for _, f := range s.RootFiles {
		if f.Name == "README.md" {
			resp, err := client.Get(f.Url)
			if err != nil {
				return "", fmt.Errorf("failed to download cookbook readme")
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			return string(body), nil
		}
	}

	return "", fmt.Errorf("failed to identify cookbook readme")
}
