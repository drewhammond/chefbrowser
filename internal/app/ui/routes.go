package ui

import (
	"crypto/tls"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/drewhammond/chefbrowser/config"
	"github.com/drewhammond/chefbrowser/internal/chef"
	"github.com/drewhammond/chefbrowser/internal/common/logging"
	"github.com/drewhammond/chefbrowser/internal/common/version"
	"github.com/foolin/goview"
	"github.com/foolin/goview/supports/ginview"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

//go:embed templates/*
var ui embed.FS

func embeddedFH(config goview.Config, tmpl string) (string, error) {
	path := filepath.Join(config.Root, tmpl)
	bytes, err := ui.ReadFile(path + config.Extension)
	return string(bytes), err
}

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

	templateRoot := "templates"
	disableCache := false
	if s.config.App.AppMode == "development" {
		s.log.Warn("development mode enabled! view cache is disabled and templates are not loaded from embed.FS")
		templateRoot = "internal/app/ui/templates"
		disableCache = true
	}

	cfg := goview.Config{
		Root:         templateRoot,
		Extension:    ".html",
		Master:       "layouts/master",
		Partials:     []string{},
		Funcs:        make(template.FuncMap),
		DisableCache: disableCache,
		Delims:       goview.Delims{Left: "{{", Right: "}}"},
	}

	cfg.Funcs["makeRunListURL"] = s.makeRunListURL
	cfg.Funcs["app_version"] = func() string { return version.Get().Version }

	gv := ginview.New(cfg)
	if s.config.App.AppMode == "production" {
		gv.ViewEngine.SetFileHandler(embeddedFH)
	}

	s.engine.HTMLRender = gv

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
		router.GET("/cookbook/:name/:version/files", s.getCookbookFiles)
		router.GET("/cookbook/:name/:version/file/*trail", s.getCookbookFile)
		router.GET("/cookbook/:name/:version/recipes", s.getCookbookRecipes)

		router.GET("/groups", s.getGroups)
		router.GET("/groups/:name", s.getGroup)

		router.GET("/policies", s.getPolicies)
		router.GET("/policy/:name", s.getPolicy)
		router.GET("/policy/:name/:revision", s.getPolicyRevision)
		router.GET("/policy-groups", s.getPolicyGroups)
		router.GET("/policy-group/:name", s.getPolicyGroup)
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
		var cookbook, recipe string
		r := strings.TrimPrefix(f, "recipe[")
		r = strings.TrimSuffix(r, "]")
		split := strings.SplitN(r, "::", 2)
		cookbook = split[0]
		if len(split) == 2 {
			recipe = split[1]
		} else {
			recipe = "default"
		}
		return fmt.Sprintf("cookbook/%s/_latest/file/recipes/%s.rb", cookbook, recipe)
	}
	if strings.HasPrefix(f, "role") {
		r := strings.TrimPrefix(f, "role[")
		r = strings.TrimSuffix(r, "]")
		return fmt.Sprintf("role/%s", r)
	}

	return ""
}

func (s *Service) getNodes(c *gin.Context) {
	query := c.Query("q")
	var nodes *chef.NodeList
	var err error
	if query != "" {
		nodes, err = s.chef.SearchNodes(c.Request.Context(), query)
	} else {
		nodes, err = s.chef.GetNodes(c.Request.Context())
	}
	if err != nil {
		s.log.Error("failed to fetch nodes", zap.Error(err))
		c.HTML(http.StatusInternalServerError, "errors/500", goview.M{
			"message": "failed to fetch nodes",
		})
		return
	}
	c.HTML(http.StatusOK, "nodes", goview.M{
		"nodes":          nodes.Nodes,
		"active_nav":     "nodes",
		"search_enabled": true,
		"title":          "All Nodes",
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
		"cookbook":   cookbook,
		"title":      cookbook.Name,
		"active_tab": "overview",
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
		"active_tab": "overview",
		"cookbook":   cookbook,
		"metadata":   metadata,
		"readme":     readme,
		"title":      cookbook.Name,
	})
}

func (s *Service) getCookbookFiles(c *gin.Context) {
	name := c.Param("name")
	version := c.Param("version")
	cookbook, err := s.chef.GetCookbookVersion(c.Request.Context(), name, version)
	if err != nil {
		c.HTML(http.StatusNotFound, "errors/404", goview.M{
			"message": "Cookbook version not found!",
		})
		return
	}
	c.HTML(http.StatusOK, "cookbook_file_list", goview.M{
		"cookbook":   cookbook,
		"active_tab": "files",
		"files":      cookbook.RootFiles,
		"title":      cookbook.Name,
	})
}

func (s *Service) getCookbookRecipes(c *gin.Context) {
	name := c.Param("name")
	version := c.Param("version")
	cookbook, err := s.chef.GetCookbookVersion(c.Request.Context(), name, version)
	if err != nil {
		c.HTML(http.StatusNotFound, "errors/404", goview.M{
			"message": "Cookbook version not found!",
		})
		return
	}
	c.HTML(http.StatusOK, "cookbook_recipes", goview.M{
		"cookbook":   cookbook,
		"active_tab": "recipes",
		"recipes":    cookbook.Recipes,
		"title":      cookbook.Name,
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

	c.HTML(http.StatusOK, "cookbook_file", goview.M{
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

func (s *Service) getPolicies(c *gin.Context) {
	policies, err := s.chef.GetPolicies(c.Request.Context())
	if err != nil {
		s.log.Warn("failed to fetch policies", zap.Error(err))
		c.HTML(http.StatusInternalServerError, "errors/500", goview.M{
			"message": "failed to fetch policies from server",
		})
		return
	}
	c.HTML(http.StatusOK, "policies", goview.M{
		"content": policies,
		"title":   "All Policies",
	})
}

func (s *Service) getPolicy(c *gin.Context) {
	name := c.Param("name")
	policy, err := s.chef.GetPolicy(c.Request.Context(), name)
	if err != nil {
		s.log.Warn("failed to fetch policy", zap.Error(err))
		c.HTML(http.StatusNotFound, "errors/404", goview.M{
			"message": "failed to fetch policy from server",
		})
		return
	}
	c.HTML(http.StatusOK, "policy", goview.M{
		"name":   name,
		"policy": policy,
		"title":  "Policy",
	})
}

func (s *Service) getPolicyRevision(c *gin.Context) {
	name := c.Param("name")
	revision := c.Param("revision")
	policy, err := s.chef.GetPolicyRevision(c.Request.Context(), name, revision)
	if err != nil {
		s.log.Warn("failed to fetch policy", zap.Error(err))
		c.HTML(http.StatusNotFound, "errors/404", goview.M{
			"message": "failed to fetch policy from server",
		})
		return
	}
	c.HTML(http.StatusOK, "policy-revision", goview.M{
		"name":     name,
		"revision": revision,
		"policy":   policy,
		"title":    "Policy",
	})
}

func (s *Service) getPolicyGroups(c *gin.Context) {
	policyGroups, err := s.chef.GetPolicyGroups(c.Request.Context())
	if err != nil {
		s.log.Warn("failed to fetch policy groups", zap.Error(err))
		c.HTML(http.StatusNotFound, "errors/404", goview.M{
			"message": "failed to fetch policy groups from server",
		})
		return
	}
	c.HTML(http.StatusOK, "policy-groups", goview.M{
		"content": policyGroups,
		"title":   "All Policy Groups",
	})
}

func (s *Service) getPolicyGroup(c *gin.Context) {
	name := c.Param("name")
	policyGroup, err := s.chef.GetPolicyGroup(c.Request.Context(), name)
	if err != nil {
		s.log.Warn("failed to fetch policy group", zap.Error(err))
		c.HTML(http.StatusNotFound, "errors/404", goview.M{
			"message": "failed to fetch policy group from server",
		})
		return
	}
	c.HTML(http.StatusOK, "policy-group", goview.M{
		"name":     name,
		"policies": policyGroup.Policies,
		"title":    "Policy group",
	})
}
