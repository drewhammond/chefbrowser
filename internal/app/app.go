package app

import (
	"fmt"
	"net"
	"net/http"
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
	engine.HidePort = true

	if cfg.App.AppMode == "development" {
		engine.Debug = true
	}

	cfg.Server.BasePath = normalizeBasePath(cfg.Server.BasePath)

	engine.Pre(middleware.RemoveTrailingSlashWithConfig(middleware.TrailingSlashConfig{
		RedirectCode: http.StatusMovedPermanently,
	}))

	engine.Use(middleware.Recover())

	if cfg.Logging.RequestLogging {
		logger.Debug("request logging is enabled")
		// DH: We might want to make these fields user configurable at some point
		logCfg := middleware.RequestLoggerConfig{
			LogLatency:      true,
			LogRemoteIP:     true,
			LogHost:         true,
			LogMethod:       true,
			LogURI:          true,
			LogUserAgent:    true,
			LogStatus:       true,
			LogResponseSize: true,
			LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
				logger.Info("request",
					zap.String("remote_ip", v.RemoteIP),
					zap.String("host", v.Host),
					zap.String("method", v.Method),
					zap.String("uri", v.URI),
					zap.Int("status", v.Status),
					zap.Int64("latency_ms", v.Latency.Milliseconds()),
					zap.Int64("response_bytes", v.ResponseSize),
					zap.String("user_agent", v.UserAgent),
				)
				return nil
			},
		}
		if !cfg.Logging.LogHealthChecks {
			logger.Debug("log_health_checks = false; requests to health check endpoint will not be logged")
			logCfg.Skipper = func(c echo.Context) bool {
				return c.Path() == cfg.Server.BasePath+"/api/health"
			}
		}
		engine.Use(middleware.RequestLoggerWithConfig(logCfg))
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
