package chef

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/drewhammond/chefbrowser/config"
	"github.com/drewhammond/chefbrowser/internal/common/logging"
	"github.com/go-chef/chef"
	"go.uber.org/zap"
)

type Interface interface {
	// Nodes
	GetNodes(ctx context.Context) (*NodeList, error)
	SearchNodes(ctx context.Context, q string) (*NodeList, error)
	GetNodesWithDetails(ctx context.Context, start, pageSize int) (*NodeListResult, error)
	SearchNodesWithDetails(ctx context.Context, q string, start, pageSize int) (*NodeListResult, error)
	GetNode(ctx context.Context, name string) (*Node, error)

	// Roles
	GetRoles(ctx context.Context) (*RoleList, error)
	GetRole(ctx context.Context, name string) (*Role, error)

	// Environments
	GetEnvironments(ctx context.Context) (interface{}, error)
	GetEnvironment(ctx context.Context, name string) (*chef.Environment, error)

	// Cookbooks
	GetCookbooks(ctx context.Context) (*CookbookListResult, error)
	GetLatestCookbooks(ctx context.Context) (*CookbookListResult, error)
	GetCookbook(ctx context.Context, name string) (*Cookbook, error)
	GetCookbookVersion(ctx context.Context, name string, version string) (*Cookbook, error)
	GetCookbookVersions(ctx context.Context, name string) ([]string, error)

	// Databags
	GetDatabags(ctx context.Context) (interface{}, error)
	GetDatabagItems(ctx context.Context, name string) (*chef.DataBagListResult, error)
	GetDatabagItemContent(ctx context.Context, databag string, item string) (chef.DataBagItem, error)

	// Policies
	GetPolicies(ctx context.Context) (chef.PoliciesGetResponse, error)
	GetPolicy(ctx context.Context, name string) (chef.PolicyGetResponse, error)
	GetPolicyRevision(ctx context.Context, name string, revision string) (chef.RevisionDetailsResponse, error)
	GetPolicyGroups(ctx context.Context) (chef.PolicyGroupGetResponse, error)
	GetPolicyGroup(ctx context.Context, name string) (PolicyGroup, error)

	// Groups
	GetGroups(ctx context.Context) (interface{}, error)
	GetGroup(ctx context.Context, name string) (chef.Group, error)
}

type Service struct {
	log    *logging.Logger
	config *config.Config
	client chef.Client
}

func New(config *config.Config, logger *logging.Logger) Interface {
	if config.App.UseMockData {
		logger.Info("Using mock data (use_mock_data = true)")
		return NewMockService(logger)
	}

	config.Chef.ServerURL = normalizeChefURL(config.Chef.ServerURL)
	logger.Info(fmt.Sprintf("initializing chef server connection (url: %s, username: %s)",
		config.Chef.ServerURL,
		config.Chef.Username))

	s := &Service{
		config: config,
		log:    logger,
	}

	key, err := os.ReadFile(config.Chef.KeyFile)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to read chef key file %s", config.Chef.KeyFile), zap.Error(err))
	}

	if !strings.HasPrefix(config.Chef.ServerURL, "https") {
		logger.Warn("Chef server connection does not use TLS. Do not use this configuration in production!")
	}

	if !config.Chef.SSLVerify {
		logger.Warn("TLS verification is disabled. Do not use this configuration in production!")
	}

	// build a client
	client, err := chef.NewClient(&chef.Config{
		Name: config.Chef.Username,
		Key:  string(key),
		// goiardi is on port 4545 by default. chef-zero is 8889
		BaseURL: config.Chef.ServerURL,
		SkipSSL: !config.Chef.SSLVerify,
	})
	if err != nil {
		logger.Fatal("failed to set up chef client", zap.Error(err))
	}

	// verify connection (we could use the global _status endpoint, but then it's not checking permissions)
	_, err = client.Nodes.List()
	if err != nil {
		logger.Fatal("failed to verify chef server connection", zap.Error(err))
	}

	s.client = *client

	return s
}

// normalizeChefURL simply adds a trailing slash to URLs to reduce confusion for users
// go-chef requires it organizations are in use.
func normalizeChefURL(url string) string {
	if strings.HasSuffix(url, "/") {
		return url
	}

	if strings.Contains(url, "/organizations/") {
		return url + "/"
	}

	return url
}
