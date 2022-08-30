package app

import (
	"fmt"
	"github.com/drewhammond/chefbrowser/config"
	"github.com/drewhammond/chefbrowser/internal/app/api"
	"github.com/drewhammond/chefbrowser/internal/app/ui"
	"github.com/drewhammond/chefbrowser/internal/chef"
	"github.com/drewhammond/chefbrowser/internal/common/logging"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AppService struct {
	Log        *logging.Logger
	Chef       *chef.Service
	UIService  *ui.Service
	APIService *api.Service
}

func New(cfg *config.Config) {
	logger := logging.New(cfg)

	logger.Info("starting chef browser",
		zap.String("version", "0.1.0"),
		zap.String("build_hash", "asdfaf"),
		zap.String("build_date", "12312312"),
	)

	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()

	engine := gin.New()
	// todo: replace with our own logger
	engine.Use(gin.Logger(), gin.Recovery())
	_ = engine.SetTrustedProxies(nil)

	chefService := chef.New(cfg, logger)

	app := AppService{
		Log:        logger,
		Chef:       chefService,
		UIService:  ui.New(cfg, engine, chefService, logger),
		APIService: api.New(cfg, engine, chefService, logger),
	}
	app.APIService.RegisterRoutes()
	//app.UIService.RegisterRoutes()
	app.UIService.RegisterRoutesWithTemplates()

	logger.Info(fmt.Sprintf("starting web server on %s", cfg.App.ListenAddr))
	err := engine.Run(cfg.App.ListenAddr)
	if err != nil {
		app.Log.Fatal("failed to start web server", zap.Error(err))
	}
}
