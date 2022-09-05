package app

import (
	"fmt"

	"github.com/drewhammond/chefbrowser/config"
	"github.com/drewhammond/chefbrowser/internal/app/api"
	"github.com/drewhammond/chefbrowser/internal/chef"
	"github.com/drewhammond/chefbrowser/internal/common/logging"
	"github.com/drewhammond/chefbrowser/internal/common/version"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AppService struct {
	Log        *logging.Logger
	Chef       *chef.Service
	APIService *api.Service
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

	engine := gin.New()
	// todo: replace with our own logger
	engine.Use(gin.Logger(), gin.Recovery())
	_ = engine.SetTrustedProxies(nil)

	chefService := chef.New(cfg, logger)

	app := AppService{
		Log:        logger,
		Chef:       chefService,
		APIService: api.New(cfg, engine, chefService, logger),
	}
	app.APIService.RegisterRoutes()

	logger.Info(fmt.Sprintf("starting web server on %s", cfg.App.ListenAddr))
	err := engine.Run(cfg.App.ListenAddr)
	if err != nil {
		app.Log.Fatal("failed to start web server", zap.Error(err))
	}
}
