package ui

import (
	"crypto/tls"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strings"

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

	cfg.Funcs["makeRunListURL"] = s.makeRunListURL

	s.engine.HTMLRender = ginview.New(cfg)

	s.engine.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/ui/nodes")
	})

	s.engine.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "errors/404", goview.M{
			"message": "Invalid route!",
		})
	})

	router := s.engine.Group("/ui")
	{
		router.GET("/", func(c *gin.Context) {
			c.Redirect(http.StatusFound, "/ui/nodes")
		})

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
		router.GET("/cookbook/:name/:version/*trail", s.getCookbookFile)

		router.GET("/groups", s.getGroups)
		router.GET("/groups/:name", s.getGroup)
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

func (s *Service) makeRunListURL(f string) string {
	if strings.HasPrefix(f, "recipe") {
		r := strings.TrimPrefix(f, "recipe[")
		r = strings.TrimSuffix(r, "]")
		split := strings.SplitN(r, "::", 2)
		return fmt.Sprintf("cookbook/%s/_latest/recipes/%s.rb", split[0], split[1])
	}
	if strings.HasPrefix(f, "role") {
		r := strings.TrimPrefix(f, "role[")
		r = strings.TrimSuffix(r, "]")
		return fmt.Sprintf("role/%s", r)
	}

	return ""
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
		c.HTML(http.StatusInternalServerError, "errors/500", goview.M{
			"message": "failed to fetch roles from server",
		})
		return
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
		if errors.Is(err, chef.ErrRoleNotFound) {
			c.HTML(http.StatusNotFound, "errors/404", goview.M{
				"message": "Role not found",
			})
			return
		}
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
		c.HTML(http.StatusNotFound, "errors/404", goview.M{
			"message": "Cookbook version not found!",
		})
		return
	}

	metadata := cookbook.Metadata

	// TODO: should we load this on the client side to speed up the initial load?
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: !s.config.Chef.SSLVerify}
	client := &http.Client{Transport: customTransport}
	readme, err := cookbook.GetReadme(c, client)
	if err != nil {
		s.log.Warn("failed to fetch cookbook", zap.Error(err))
	}
	c.HTML(http.StatusOK, "cookbook", goview.M{
		"cookbook": cookbook,
		"metadata": metadata,
		"readme":   readme,
		"title":    cookbook.Name,
	})
}

func (s *Service) getCookbookFile(c *gin.Context) {
	name := c.Param("name")
	version := c.Param("version")
	// *trail always contains a leading slash apparently
	path := strings.TrimPrefix(c.Param("trail"), "/")
	cookbook, err := s.chef.GetCookbookVersion(c.Request.Context(), name, version)
	if err != nil {
		c.HTML(http.StatusNotFound, "errors/404", goview.M{
			"message": "Cookbook version not found!",
		})
		return
	}

	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: !s.config.Chef.SSLVerify}
	client := &http.Client{Transport: customTransport}
	file, err := cookbook.GetFile(c, client, path)
	if err != nil {
		s.log.Warn("failed to fetch cookbook", zap.Error(err))
		c.HTML(http.StatusNotFound, "errors/404", goview.M{
			"message": "Cookbook file not found!",
		})
		return
	}

	c.HTML(http.StatusOK, "cookbook", goview.M{
		"cookbook":   cookbook,
		"active_tab": "files",
		"file":       file,
		"path":       path,
		"title":      cookbook.Name,
	})
}

func (s *Service) getCookbooks(c *gin.Context) {
	cookbooks, err := s.chef.GetCookbooks(c.Request.Context())
	if err != nil {
		s.log.Warn("failed to fetch cookbooks", zap.Error(err))
		c.HTML(http.StatusInternalServerError, "errors/500", goview.M{
			"message": "failed to fetch cookbooks from server",
		})
		return
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
		c.HTML(http.StatusInternalServerError, "errors/500", goview.M{
			"message": "failed to fetch environments from server",
		})
		return
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
		if errors.Is(err, chef.ErrEnvironmentNotFound) {
			c.HTML(http.StatusNotFound, "errors/404", goview.M{
				"message": "Environment not found",
			})
			return
		}
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
		c.HTML(http.StatusInternalServerError, "errors/500", goview.M{
			"message": "failed to fetch databags from server",
		})
		return
	}
	c.HTML(http.StatusOK, "databags", goview.M{
		"databags": databags,
		"title":    "Showing all data bags",
	})
}

func (s *Service) getDatabagItems(c *gin.Context) {
	name := c.Param("name")
	items, err := s.chef.GetDatabagItems(c.Request.Context(), name)
	if err != nil {
		if errors.Is(err, chef.ErrDatabagNotFound) {
			s.log.Warn("failed to fetch databag items", zap.Error(err))

			c.HTML(http.StatusNotFound, "errors/404", goview.M{
				"message": "Databag not found",
			})
			return
		}
	}
	c.HTML(http.StatusOK, "databag_items", goview.M{
		"databag": name,
		"items":   items,
		"title":   "Data Bag Items",
	})
}

func (s *Service) getDatabagItemContent(c *gin.Context) {
	databag := c.Param("name")
	item := c.Param("item")
	content, err := s.chef.GetDatabagItemContent(c.Request.Context(), databag, item)
	if err != nil {
		if errors.Is(err, chef.ErrDatabagItemNotFound) {
			s.log.Warn("failed to fetch databag item content", zap.Error(err))
			c.HTML(http.StatusNotFound, "errors/404", goview.M{
				"message": "Databag item not found",
			})
			return
		}
	}
	c.HTML(http.StatusOK, "databag_item_content", goview.M{
		"databag": databag,
		"item":    item,
		"content": content,
		"title":   "Data Bag Items",
	})
}

func (s *Service) getGroups(c *gin.Context) {
	groups, err := s.chef.GetGroups(c.Request.Context())
	if err != nil {
		s.log.Warn("failed to fetch groups", zap.Error(err))
		c.HTML(http.StatusInternalServerError, "errors/500", goview.M{
			"message": "failed to fetch groups from server",
		})
		return
	}
	c.HTML(http.StatusOK, "groups", goview.M{
		"content": groups,
		"title":   "All Groups",
	})
}

func (s *Service) getGroup(c *gin.Context) {
	name := c.Param("name")
	group, err := s.chef.GetGroup(c.Request.Context(), name)
	if err != nil {
		s.log.Warn("failed to fetch group", zap.Error(err))
		c.HTML(http.StatusNotFound, "errors/404", goview.M{
			"message": "failed to fetch group from server",
		})
		return
	}
	c.HTML(http.StatusOK, "group", goview.M{
		"content": group,
		"title":   "All Groups",
	})
}
