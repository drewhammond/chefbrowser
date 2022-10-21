package app

import (
	"fmt"

	"github.com/drewhammond/chefbrowser/config"
	"github.com/drewhammond/chefbrowser/internal/app/api"
	"github.com/drewhammond/chefbrowser/internal/app/ui"
	"github.com/drewhammond/chefbrowser/internal/chef"
	"github.com/drewhammond/chefbrowser/internal/common/logging"
	"github.com/drewhammond/chefbrowser/internal/common/version"
	"github.com/gin-gonic/gin"
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

	if cfg.App.AppMode == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := echo.New()

	engine.Use(middleware.Recover())

	if cfg.Logging.RequestLogging {
		// todo: replace with our own logger
		engine.Use(middleware.Logger())
	}

	if cfg.Server.EnableGzip {
		engine.Use(middleware.Gzip())
	}

	//if cfg.Server.TrustedProxies == "" {
	//	_ = engine.SetTrustedProxies(nil)
	//} else {
	//	_ = engine.SetTrustedProxies(strings.Split(cfg.Server.TrustedProxies, ","))
	//}

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
