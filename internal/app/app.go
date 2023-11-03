package app

import (
	"fmt"
	"net"
	"path"
	"strings"

	"github.com/drewhammond/chefbrowser/config"
	"github.com/drewhammond/chefbrowser/internal/app/api"
	"github.com/drewhammond/chefbrowser/internal/app/ui"
	"github.com/drewhammond/chefbrowser/internal/chef"
	"github.com/drewhammond/chefbrowser/internal/common/logging"
	"github.com/drewhammond/chefbrowser/internal/common/version"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

type AppService struct {
	Log        *logging.Logger
	Chef       *chef.Service
	APIService *api.Service
	UIService  *ui.Service
}

func New(cfg *config.Config) {
	logger := logging.New(cfg)

	logger.Info("starting chef browser",
		zap.String("version", version.Get().Version),
		zap.String("build_hash", version.Get().BuildHash),
		zap.String("build_date", version.Get().BuildDate),
	)

	engine := echo.New()
	engine.HideBanner = true

	if cfg.App.AppMode == "development" {
		engine.Debug = true
	}

	engine.Pre(middleware.RemoveTrailingSlash())

	engine.Use(middleware.Recover())

	if cfg.Logging.RequestLogging {
		// todo: replace with our own logger
		engine.Use(middleware.Logger())
	}

	if cfg.Server.EnableGzip {
		engine.Use(middleware.Gzip())
	}

	if cfg.Server.TrustedProxies != "" {
		var opts []echo.TrustOption
		opts = append(opts, echo.TrustPrivateNet(false))
		for _, x := range strings.Split(cfg.Server.TrustedProxies, ",") {
			_, j, err := net.ParseCIDR(x)
			if err != nil {
				logger.Error(fmt.Sprintf("invalid proxy network specified: %s", x))
				continue
			}
			opts = append(opts, echo.TrustIPRange(j))
		}
		engine.IPExtractor = echo.ExtractIPFromXFFHeader(opts...)
	} else {
		engine.IPExtractor = echo.ExtractIPDirect()
	}

	chefService := chef.New(cfg, logger)

	cfg.Server.BasePath = normalizeBasePath(cfg.Server.BasePath)

	app := AppService{
		Log:        logger,
		Chef:       chefService,
		APIService: api.New(cfg, engine, chefService, logger),
		UIService:  ui.New(cfg, engine, chefService, logger),
	}
	app.APIService.RegisterRoutes()
	app.UIService.RegisterRoutes()

	logger.Info(fmt.Sprintf("starting web server on %s", cfg.App.ListenAddr))
	err := engine.Start(cfg.App.ListenAddr)
	if err != nil {
		app.Log.Fatal("failed to start web server", zap.Error(err))
	}
}

// normalizeBasePath cleans and strips trailing slashes from the configured base_path
func normalizeBasePath(p string) string {
	p = path.Clean(p)
	if p == "." || p == "/" {
		return ""
	}
	return p
}
