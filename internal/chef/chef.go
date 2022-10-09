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
	GetCookbook(ctx context.Context) (Cookbook, error)
	GetCookbooks(ctx context.Context) ([]Cookbook, error)
}

type Service struct {
	Interface
	log    *logging.Logger
	config *config.Config
	client chef.Client
}

func (s Service) GetClient() *chef.Client {
	return &s.client
}

func New(config *config.Config, logger *logging.Logger) *Service {
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
	// TODO: better health check? move out of the constructor?
	_, err = client.Clients.List()
	if err != nil {
		logger.Fatal("failed to verify chef server connection", zap.Error(err))
	}

	s.client = *client

	return s
}
