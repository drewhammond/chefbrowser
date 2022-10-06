package ui

import (
	"html/template"
	"net/http"

	"github.com/drewhammond/chefbrowser/config"
	"github.com/drewhammond/chefbrowser/internal/chef"
	"github.com/drewhammond/chefbrowser/internal/common/logging"
	"github.com/foolin/goview"
	"github.com/foolin/goview/supports/ginview"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Service struct {
	log    *logging.Logger
	config *config.Config
	chef   *chef.Service
	engine *gin.Engine
}

func New(config *config.Config, engine *gin.Engine, chef *chef.Service, logger *logging.Logger) *Service {
	s := Service{
		config: config,
		chef:   chef,
		log:    logger,
		engine: engine,
	}
	return &s
}

func (s *Service) RegisterRoutes() {
	s.log.Info("registering UI routes")

	cfg := goview.Config{
		Root:         "templates",
		Extension:    ".html",
		Master:       "layouts/master",
		Partials:     []string{},
		Funcs:        make(template.FuncMap),
		DisableCache: true,
		Delims:       goview.Delims{Left: "{{", Right: "}}"},
	}
	s.engine.HTMLRender = ginview.New(cfg)

	router := s.engine.Group("/ui")
	{
		router.GET("/nodes", s.getNodes)
		router.GET("/node/:name", s.getNode)

		router.GET("/environments", s.getEnvironments)
		router.GET("/environment/:name", s.getEnvironment)

		router.GET("/roles", s.getRoles)
		router.GET("/role/:name", s.getRole)

		router.GET("/databags", s.getDatabags)
		router.GET("/databag/:name", s.getDatabagItems)
		router.GET("/databag/:name/:item", s.getDatabagItemContent)

		router.GET("/cookbooks", s.getCookbooks)
		router.GET("/cookbook/:name", s.getCookbook)
		router.GET("/cookbook/:name/:version", s.getCookbookVersion)

		// router.GET("/groups", s.getGroups)
		// router.GET("/groups/:name", s.getGroup)
	}
}

func (s *Service) getNode(c *gin.Context) {
	name := c.Param("name")
	node, err := s.chef.GetNode(c.Request.Context(), name)
	if err != nil {
		s.log.Error("failed to fetch node from server", zap.Error(err))
	}

	c.HTML(http.StatusOK, "node", gin.H{
		"node":  node,
		"title": node.Name,
	})
}

func (s *Service) getNodes(c *gin.Context) {
	nodes, err := s.chef.GetNodes(c.Request.Context())
	if err != nil {
		s.log.Error("failed to fetch nodes", zap.Error(err))
	}
	c.HTML(http.StatusOK, "nodes", goview.M{
		"nodes": nodes.Nodes,
		"title": "All Nodes",
	})
}

func (s *Service) getRoles(c *gin.Context) {
	roles, err := s.chef.GetRoles(c.Request.Context())
	if err != nil {
		s.log.Error("failed to fetch roles", zap.Error(err))
	}
	c.HTML(http.StatusOK, "roles", goview.M{
		"roles": roles.Roles,
		"title": "All Nodes",
	})
}

func (s *Service) getRole(c *gin.Context) {
	name := c.Param("name")
	role, err := s.chef.GetRole(c.Request.Context(), name)
	if err != nil {
		s.log.Warn("failed to fetch role", zap.Error(err))
	}
	c.HTML(http.StatusOK, "role", goview.M{
		"role":  role,
		"title": role.Name,
	})
}

func (s *Service) getCookbook(c *gin.Context) {
	name := c.Param("name")
	cookbook, err := s.chef.GetCookbook(c.Request.Context(), name)
	if err != nil {
		s.log.Warn("failed to fetch cookbook", zap.Error(err))
	}
	c.HTML(http.StatusOK, "cookbook", goview.M{
		"cookbook": cookbook,
		"title":    cookbook.Name,
	})
}

func (s *Service) getCookbookVersion(c *gin.Context) {
	name := c.Param("name")
	version := c.Param("version")
	cookbook, err := s.chef.GetCookbookVersion(c.Request.Context(), name, version)
	if err != nil {
		s.log.Warn("failed to fetch cookbook", zap.Error(err))
	}
	c.HTML(http.StatusOK, "cookbook", goview.M{
		"cookbook": cookbook,
		"title":    cookbook.Name,
	})
}

func (s *Service) getCookbooks(c *gin.Context) {
	cookbooks, err := s.chef.GetCookbooks(c.Request.Context())
	if err != nil {
		s.log.Warn("failed to fetch cookbooks", zap.Error(err))
	}
	c.HTML(http.StatusOK, "cookbooks", goview.M{
		"cookbooks": cookbooks.Cookbooks,
		"title":     "All Cookbooks",
	})
}

func (s *Service) getEnvironments(c *gin.Context) {
	environments, err := s.chef.GetEnvironments(c.Request.Context())
	if err != nil {
		s.log.Warn("failed to fetch environments", zap.Error(err))
	}
	c.HTML(http.StatusOK, "environments", goview.M{
		"environments": environments,
		"title":        "All Environments",
	})
}

func (s *Service) getEnvironment(c *gin.Context) {
	name := c.Param("name")
	environment, err := s.chef.GetEnvironment(c.Request.Context(), name)
	if err != nil {
		s.log.Warn("failed to fetch environment", zap.Error(err))
	}
	c.HTML(http.StatusOK, "environment", goview.M{
		"environment": environment,
		"title":       environment.Name,
	})
}

func (s *Service) getDatabags(c *gin.Context) {
	databags, err := s.chef.GetDatabags(c.Request.Context())
	if err != nil {
		s.log.Warn("failed to fetch databags", zap.Error(err))
	}
	c.HTML(http.StatusOK, "databags", goview.M{
		"databags": databags,
		"title":    "Showing all data bags",
	})
}

func (s *Service) getDatabagItems(c *gin.Context) {
	name := c.Param("name")
	databagItems, err := s.chef.GetDatabagItems(c.Request.Context(), name)
	if err != nil {
		s.log.Warn("failed to fetch databag items", zap.Error(err))
	}
	c.HTML(http.StatusOK, "databag", goview.M{
		"items": databagItems,
		"title": "Data Bag Items",
	})
}

func (s *Service) getDatabagItemContent(c *gin.Context) {
	databag := c.Param("name")
	item := c.Param("item")
	content, err := s.chef.GetDatabagItemContent(c.Request.Context(), databag, item)
	if err != nil {
		s.log.Warn("failed to fetch databag item content", zap.Error(err))
	}
	c.HTML(http.StatusOK, "databag_item_content", goview.M{
		"content": content,
		"title":   "Data Bag Items",
	})
}
