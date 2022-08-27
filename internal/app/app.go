package app

import (
	"fmt"
	"net/http"

	"github.com/drewhammond/chefbrowser/config"
	"github.com/drewhammond/chefbrowser/internal/chef"
	"github.com/drewhammond/chefbrowser/internal/common/logging"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AppService struct {
	Log  *logging.Logger
	Chef *chef.Service
}

func (r *AppService) getNode(c *gin.Context) {
	name := c.Param("name")
	r.Log.Debug("getting node from chef server")
	node, err := r.Chef.GetNode(c.Request.Context(), name)
	if err != nil {
		r.Log.Error("failed to fetch node from server", zap.Error(err))
	}
	c.JSON(http.StatusOK, node)
}

func (r *AppService) getNodes(c *gin.Context) {
	r.Log.Debug("getting all nodes from chef server")
	nodes, err := r.Chef.GetNodes(c.Request.Context())
	if err != nil {
		r.Log.Error("failed to fetch nodes", zap.Error(err))
	}
	c.JSON(http.StatusOK, nodes)
}

func (r *AppService) getRoles(c *gin.Context) {
	r.Log.Debug("getting all roles from chef server")
	roles, err := r.Chef.GetRoles(c.Request.Context())
	if err != nil {
		r.Log.Error("failed to fetch roles from server", zap.Error(err))
	}
	c.JSON(http.StatusOK, roles)
}

func (r *AppService) getRole(c *gin.Context) {
	name := c.Param("name")
	r.Log.Debug("getting role from chef server")
	role, err := r.Chef.GetRole(c.Request.Context(), name)
	if err != nil {
		r.Log.Error("failed to fetch role from server", zap.Error(err))
	}
	c.JSON(http.StatusOK, role)
}

func (r *AppService) getEnvironments(c *gin.Context) {
	r.Log.Debug("getting all environments from chef server")
	environments, err := r.Chef.GetEnvironments(c.Request.Context())
	if err != nil {
		r.Log.Error("failed to fetch environments from server", zap.Error(err))
	}
	c.JSON(http.StatusOK, environments)
}

func (r *AppService) getEnvironment(c *gin.Context) {
	name := c.Param("name")
	r.Log.Debug(fmt.Sprintf("getting environment %s from chef server", name))
	environment, err := r.Chef.GetEnvironment(c.Request.Context(), name)
	if err != nil {
		r.Log.Error(fmt.Sprintf("failed to fetch environment %s from server", name), zap.Error(err))
	}
	if environment != nil {
		c.JSON(http.StatusOK, environment)
		return
	}

	c.JSON(http.StatusNotFound, environment)
}

func (r *AppService) getCookbooks(c *gin.Context) {
	r.Log.Debug("getting all cookbooks from chef server")
	cookbooks, err := r.Chef.GetCookbooks(c.Request.Context())
	if err != nil {
		r.Log.Error("failed to fetch cookbooks from server", zap.Error(err))
	}
	c.JSON(http.StatusOK, cookbooks)
}

func (r *AppService) getCookbook(c *gin.Context) {
	name := c.Param("name")
	cookbook, err := r.Chef.GetCookbook(c.Request.Context(), name)
	if err != nil {
		r.Log.Error("failed to fetch cookbook from server", zap.Error(err))
	}
	c.JSON(http.StatusOK, cookbook)
}

func (r *AppService) getCookbookVersion(c *gin.Context) {
	name := c.Param("name")
	version := c.Param("version")
	cookbook, err := r.Chef.GetCookbookVersion(c.Request.Context(), name, version)
	if err != nil {
		r.Log.Error("failed to fetch cookbook from server", zap.Error(err))
	}
	c.JSON(http.StatusOK, cookbook)
}

func (r *AppService) getGroups(c *gin.Context) {
	groups, err := r.Chef.GetGroups(c.Request.Context())
	if err != nil {
		r.Log.Error("failed to fetch groups from server", zap.Error(err))
	}
	c.JSON(http.StatusOK, groups)
}

func (r *AppService) getGroup(c *gin.Context) {
	name := c.Param("name")
	group, err := r.Chef.GetGroup(c.Request.Context(), name)
	if err != nil {
		r.Log.Error("failed to fetch group from server", zap.Error(err))
	}
	c.JSON(http.StatusOK, group)
}

type HealthResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func getHealth(c *gin.Context) {
	c.JSON(http.StatusOK, &HealthResponse{Success: true, Message: "ready"})
}

type ConfigService interface {
	Get()
}

func New(cfg *config.Config) {
	logger := logging.New(cfg)

	logger.Info("starting chef browser",
		zap.String("version", "0.1.0"),
		zap.String("build_hash", "asdfaf"),
		zap.String("build_date", "12312312"),
	)

	app := AppService{
		Log:  logger,
		Chef: chef.New(cfg, logger),
	}

	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()

	router := gin.New()
	// todo: replace with our own logger
	router.Use(gin.Logger(), gin.Recovery())

	_ = router.SetTrustedProxies(nil)

	api := router.Group("/api")
	{
		// nodes
		api.GET("/nodes", app.getNodes)
		api.GET("/node/:name", app.getNode)

		// environments
		api.GET("/environments", app.getEnvironments)
		api.GET("/environment/:name", app.getEnvironment)

		// roles
		api.GET("/roles", app.getRoles)
		api.GET("/role/:name", app.getRole)

		// cookbooks
		api.GET("/cookbooks", app.getCookbooks)
		api.GET("/cookbooks/:name", app.getCookbook)
		api.GET("/cookbooks/:name/:version", app.getCookbookVersion)

		// groups
		api.GET("/groups", app.getGroups)
		api.GET("/groups/:name", app.getGroup)

		// misc
		api.GET("/health", getHealth)
	}

	logger.Info(fmt.Sprintf("starting web server on %s", cfg.App.ListenAddr))
	err := router.Run(cfg.App.ListenAddr)
	if err != nil {
		app.Log.Fatal("failed to start web server", zap.Error(err))
	}
}
