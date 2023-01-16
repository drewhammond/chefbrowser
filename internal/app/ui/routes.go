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
	"github.com/foolin/goview/supports/echoview-v4"
	"github.com/labstack/echo/v4"
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
	engine *echo.Echo
}

func New(config *config.Config, engine *echo.Echo, chef *chef.Service, logger *logging.Logger) *Service {
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

	ev := echoview.New(cfg)
	if s.config.App.AppMode == "production" {
		ev.ViewEngine.SetFileHandler(embeddedFH)
	}

	s.engine.Renderer = ev

	s.engine.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusFound, "/ui/nodes")
	})

	s.engine.RouteNotFound("/*", func(c echo.Context) error {
		return c.Render(http.StatusNotFound, "errors/404", echo.Map{
			"message": "Invalid route!",
		})
	})

	router := s.engine.Group("/ui")
	{
		router.GET("/", func(c echo.Context) error {
			return c.Redirect(http.StatusFound, "/ui/nodes")
		})
		router.GET("/nodes", s.getNodes)
		router.GET("/nodes/:name", s.getNode)

		router.GET("/environments", s.getEnvironments)
		router.GET("/environments/:name", s.getEnvironment)

		router.GET("/roles", s.getRoles)
		router.GET("/roles/:name", s.getRole)

		router.GET("/databags", s.getDatabags)
		router.GET("/databags/:name", s.getDatabagItems)
		router.GET("/databags/:name/:item", s.getDatabagItemContent)

		router.GET("/cookbooks", s.getCookbooks)
		router.GET("/cookbooks/:name", s.getCookbook)
		router.GET("/cookbooks/:name/:version", s.getCookbookVersion)
		router.GET("/cookbooks/:name/:version/files", s.getCookbookFiles)
		router.GET("/cookbooks/:name/:version/file/*", s.getCookbookFile)
		router.GET("/cookbooks/:name/:version/recipes", s.getCookbookRecipes)

		router.GET("/groups", s.getGroups)
		router.GET("/groups/:name", s.getGroup)

		router.GET("/policies", s.getPolicies)
		router.GET("/policies/:name", s.getPolicy)
		router.GET("/policies/:name/:revision", s.getPolicyRevision)
		router.GET("/policy-groups", s.getPolicyGroups)
		router.GET("/policy-groups/:name", s.getPolicyGroup)
	}
}

func (s *Service) getNode(c echo.Context) error {
	name := c.Param("name")
	node, err := s.chef.GetNode(c.Request().Context(), name)
	if err != nil {
		return c.Render(http.StatusNotFound, "errors/404", echo.Map{
			"message": "Node not found",
		})
	}

	return c.Render(http.StatusOK, "node", echo.Map{
		"active_nav": "nodes",
		"node":       node,
		"title":      node.Name,
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
		return fmt.Sprintf("cookbooks/%s/_latest/file/recipes/%s.rb", cookbook, recipe)
	}
	if strings.HasPrefix(f, "role") {
		r := strings.TrimPrefix(f, "role[")
		r = strings.TrimSuffix(r, "]")
		return fmt.Sprintf("roles/%s", r)
	}

	return ""
}

func (s *Service) getNodes(c echo.Context) error {
	query := c.QueryParam("q")
	var nodes *chef.NodeList
	var err error
	if query != "" {
		nodes, err = s.chef.SearchNodes(c.Request().Context(), query)
	} else {
		nodes, err = s.chef.GetNodes(c.Request().Context())
	}
	if err != nil {
		s.log.Error("failed to fetch nodes", zap.Error(err))
		return c.Render(http.StatusInternalServerError, "errors/500", echo.Map{
			"message": "failed to fetch nodes",
		})

	}
	return c.Render(http.StatusOK, "nodes", echo.Map{
		"nodes":          nodes.Nodes,
		"active_nav":     "nodes",
		"search_enabled": true,
		"title":          "All Nodes",
	})
}

func (s *Service) getRoles(c echo.Context) error {
	roles, err := s.chef.GetRoles(c.Request().Context())
	if err != nil {
		s.log.Error("failed to fetch roles", zap.Error(err))
		return c.Render(http.StatusInternalServerError, "errors/500", echo.Map{
			"message": "failed to fetch roles from server",
		})

	}
	return c.Render(http.StatusOK, "roles", echo.Map{
		"roles":      roles.Roles,
		"active_nav": "roles",
		"title":      "All Roles",
	})
}

func (s *Service) getRole(c echo.Context) error {
	name := c.Param("name")
	role, err := s.chef.GetRole(c.Request().Context(), name)
	if err != nil {
		if errors.Is(err, chef.ErrRoleNotFound) {
			return c.Render(http.StatusNotFound, "errors/404", echo.Map{
				"message": "Role not found",
			})
		}
	}
	return c.Render(http.StatusOK, "role", echo.Map{
		"role":       role,
		"active_nav": "roles",
		"title":      role.Name,
	})
}

func (s *Service) getCookbook(c echo.Context) error {
	name := c.Param("name")
	cookbook, err := s.chef.GetCookbook(c.Request().Context(), name)
	if err != nil {
		s.log.Warn("failed to fetch cookbook", zap.Error(err))
	}
	return c.Render(http.StatusOK, "cookbook", echo.Map{
		"cookbook":   cookbook,
		"title":      cookbook.Name,
		"active_nav": "cookbooks",
		"active_tab": "overview",
	})
}

func (s *Service) getCookbookVersion(c echo.Context) error {
	name := c.Param("name")
	version := c.Param("version")
	cookbook, err := s.chef.GetCookbookVersion(c.Request().Context(), name, version)
	if err != nil {
		if errors.Is(err, chef.ErrCookbookVersionNotFound) {
			return c.Render(http.StatusNotFound, "errors/404", echo.Map{
				"message": "Cookbook version not found!",
			})
		}
		return c.Render(http.StatusInternalServerError, "errors/500", echo.Map{
			"message": "Unknown error occurred",
		})
	}

	metadata := cookbook.Metadata

	// TODO: should we load this on the client side to speed up the initial load?
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: !s.config.Chef.SSLVerify}
	client := &http.Client{Transport: customTransport}
	readme, err := cookbook.GetReadme(c.Request().Context(), client)
	if err != nil {
		s.log.Warn("failed to fetch cookbook", zap.Error(err))
	}
	return c.Render(http.StatusOK, "cookbook", echo.Map{
		"active_tab": "overview",
		"active_nav": "cookbooks",
		"cookbook":   cookbook,
		"metadata":   metadata,
		"readme":     readme,
		"title":      cookbook.Name,
	})
}

func (s *Service) getCookbookFiles(c echo.Context) error {
	name := c.Param("name")
	version := c.Param("version")
	cookbook, err := s.chef.GetCookbookVersion(c.Request().Context(), name, version)
	if err != nil {
		return c.Render(http.StatusNotFound, "errors/404", echo.Map{
			"message": "Cookbook version not found!",
		})
	}
	return c.Render(http.StatusOK, "cookbook_file_list", echo.Map{
		"cookbook":   cookbook,
		"active_tab": "files",
		"active_nav": "cookbooks",
		"files":      cookbook.RootFiles,
		"title":      cookbook.Name,
	})
}

func (s *Service) getCookbookRecipes(c echo.Context) error {
	name := c.Param("name")
	version := c.Param("version")
	cookbook, err := s.chef.GetCookbookVersion(c.Request().Context(), name, version)
	if err != nil {
		return c.Render(http.StatusNotFound, "errors/404", echo.Map{
			"message": "Cookbook version not found!",
		})
	}
	return c.Render(http.StatusOK, "cookbook_recipes", echo.Map{
		"cookbook":   cookbook,
		"active_tab": "recipes",
		"active_nav": "cookbooks",
		"recipes":    cookbook.Recipes,
		"title":      cookbook.Name,
	})
}

func (s *Service) getCookbookFile(c echo.Context) error {
	name := c.Param("name")
	version := c.Param("version")
	path := c.Param("*")
	cookbook, err := s.chef.GetCookbookVersion(c.Request().Context(), name, version)
	if err != nil {
		return c.Render(http.StatusNotFound, "errors/404", echo.Map{
			"message": "Cookbook version not found!",
		})
	}

	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: !s.config.Chef.SSLVerify}
	client := &http.Client{Transport: customTransport}
	file, err := cookbook.GetFile(c.Request().Context(), client, path)
	if err != nil {
		s.log.Warn("failed to fetch cookbook", zap.Error(err))
		return c.Render(http.StatusNotFound, "errors/404", echo.Map{
			"message": "Cookbook file not found!",
		})
	}

	return c.Render(http.StatusOK, "cookbook_file", echo.Map{
		"cookbook":   cookbook,
		"active_tab": "files",
		"active_nav": "cookbooks",
		"file":       file,
		"path":       path,
		"title":      cookbook.Name,
	})
}

func (s *Service) getCookbooks(c echo.Context) error {
	cookbooks, err := s.chef.GetCookbooks(c.Request().Context())
	if err != nil {
		s.log.Warn("failed to fetch cookbooks", zap.Error(err))
		return c.Render(http.StatusInternalServerError, "errors/500", echo.Map{
			"message": "failed to fetch cookbooks from server",
		})
	}

	return c.Render(http.StatusOK, "cookbooks", echo.Map{
		"cookbooks":  cookbooks.Cookbooks,
		"active_nav": "cookbooks",
		"title":      "All Cookbooks",
	})
}

func (s *Service) getEnvironments(c echo.Context) error {
	environments, err := s.chef.GetEnvironments(c.Request().Context())
	if err != nil {
		s.log.Warn("failed to fetch environments", zap.Error(err))
		return c.Render(http.StatusInternalServerError, "errors/500", echo.Map{
			"message": "failed to fetch environments from server",
		})
	}
	return c.Render(http.StatusOK, "environments", echo.Map{
		"environments": environments,
		"active_nav":   "environments",
		"title":        "All Environments",
	})
}

func (s *Service) getEnvironment(c echo.Context) error {
	name := c.Param("name")
	environment, err := s.chef.GetEnvironment(c.Request().Context(), name)
	if err != nil {
		s.log.Warn("failed to fetch environment", zap.Error(err))
		if errors.Is(err, chef.ErrEnvironmentNotFound) {
			return c.Render(http.StatusNotFound, "errors/404", echo.Map{
				"message": "Environment not found",
			})
		}
	}
	return c.Render(http.StatusOK, "environment", echo.Map{
		"environment": environment,
		"active_nav":  "environments",
		"title":       environment.Name,
	})
}

func (s *Service) getDatabags(c echo.Context) error {
	databags, err := s.chef.GetDatabags(c.Request().Context())
	if err != nil {
		s.log.Warn("failed to fetch databags", zap.Error(err))
		return c.Render(http.StatusInternalServerError, "errors/500", echo.Map{
			"message": "failed to fetch databags from server",
		})
	}
	return c.Render(http.StatusOK, "databags", echo.Map{
		"databags":   databags,
		"active_nav": "databags",
		"title":      "All Data Bags",
	})
}

func (s *Service) getDatabagItems(c echo.Context) error {
	name := c.Param("name")
	items, err := s.chef.GetDatabagItems(c.Request().Context(), name)
	if err != nil {
		if errors.Is(err, chef.ErrDatabagNotFound) {
			s.log.Warn("failed to fetch databag items", zap.Error(err))

			return c.Render(http.StatusNotFound, "errors/404", echo.Map{
				"message": "Databag not found",
			})
		}
	}
	return c.Render(http.StatusOK, "databag_items", echo.Map{
		"databag":    name,
		"items":      items,
		"active_nav": "databags",
		"title":      fmt.Sprintf("Data Bag %s - All Items", name),
	})
}

func (s *Service) getDatabagItemContent(c echo.Context) error {
	databag := c.Param("name")
	item := c.Param("item")
	content, err := s.chef.GetDatabagItemContent(c.Request().Context(), databag, item)
	if err != nil {
		if errors.Is(err, chef.ErrDatabagItemNotFound) {
			s.log.Warn("failed to fetch databag item content", zap.Error(err))
			return c.Render(http.StatusNotFound, "errors/404", echo.Map{
				"message": "Databag item not found",
			})
		}
	}
	return c.Render(http.StatusOK, "databag_item_content", echo.Map{
		"active_nav": "databags",
		"databag":    databag,
		"item":       item,
		"content":    content,
		"title":      fmt.Sprintf("Data Bag %s - %s", databag, item),
	})
}

func (s *Service) getGroups(c echo.Context) error {
	groups, err := s.chef.GetGroups(c.Request().Context())
	if err != nil {
		s.log.Warn("failed to fetch groups", zap.Error(err))
		return c.Render(http.StatusInternalServerError, "errors/500", echo.Map{
			"message": "failed to fetch groups from server",
		})
	}
	return c.Render(http.StatusOK, "groups", echo.Map{
		"content":    groups,
		"active_nav": "groups",
		"title":      "All Groups",
	})
}

func (s *Service) getGroup(c echo.Context) error {
	name := c.Param("name")
	group, err := s.chef.GetGroup(c.Request().Context(), name)
	if err != nil {
		s.log.Warn("failed to fetch group", zap.Error(err))
		return c.Render(http.StatusNotFound, "errors/404", echo.Map{
			"message": "failed to fetch group from server",
		})
	}
	return c.Render(http.StatusOK, "group", echo.Map{
		"content":    group,
		"active_nav": "groups",
		"title":      fmt.Sprintf("Groups - %s", name),
	})
}

func (s *Service) getPolicies(c echo.Context) error {
	policies, err := s.chef.GetPolicies(c.Request().Context())
	if err != nil {
		s.log.Warn("failed to fetch policies", zap.Error(err))
		return c.Render(http.StatusInternalServerError, "errors/500", echo.Map{
			"message": "failed to fetch policies from server",
		})
	}
	return c.Render(http.StatusOK, "policies", echo.Map{
		"content":    policies,
		"active_nav": "policies",
		"title":      "All Policies",
	})
}

func (s *Service) getPolicy(c echo.Context) error {
	name := c.Param("name")
	policy, err := s.chef.GetPolicy(c.Request().Context(), name)
	if err != nil {
		s.log.Warn("failed to fetch policy", zap.Error(err))
		return c.Render(http.StatusNotFound, "errors/404", echo.Map{
			"message": "failed to fetch policy from server",
		})
	}
	return c.Render(http.StatusOK, "policy", echo.Map{
		"name":       name,
		"policy":     policy,
		"active_nav": "policies",
		"title":      fmt.Sprintf("Policy > %s", name),
	})
}

func (s *Service) getPolicyRevision(c echo.Context) error {
	name := c.Param("name")
	revision := c.Param("revision")
	policy, err := s.chef.GetPolicyRevision(c.Request().Context(), name, revision)
	if err != nil {
		s.log.Warn("failed to fetch policy", zap.Error(err))
		return c.Render(http.StatusNotFound, "errors/404", echo.Map{
			"message": "failed to fetch policy from server",
		})
	}
	return c.Render(http.StatusOK, "policy-revision", echo.Map{
		"active_nav": "policies",
		"name":       name,
		"revision":   revision,
		"policy":     policy,
		"title":      fmt.Sprintf("%s > %s", name, revision),
	})
}

func (s *Service) getPolicyGroups(c echo.Context) error {
	policyGroups, err := s.chef.GetPolicyGroups(c.Request().Context())
	if err != nil {
		s.log.Warn("failed to fetch policy groups", zap.Error(err))
		return c.Render(http.StatusNotFound, "errors/404", echo.Map{
			"message": "failed to fetch policy groups from server",
		})
	}
	return c.Render(http.StatusOK, "policy-groups", echo.Map{
		"content":    policyGroups,
		"active_nav": "policies",
		"title":      "All Policy Groups",
	})
}

func (s *Service) getPolicyGroup(c echo.Context) error {
	name := c.Param("name")
	policyGroup, err := s.chef.GetPolicyGroup(c.Request().Context(), name)
	if err != nil {
		s.log.Warn("failed to fetch policy group", zap.Error(err))
		return c.Render(http.StatusNotFound, "errors/404", echo.Map{
			"message": "failed to fetch policy group from server",
		})
	}
	return c.Render(http.StatusOK, "policy-group", echo.Map{
		"active_nav": "policies",
		"name":       name,
		"policies":   policyGroup.Policies,
		"title":      fmt.Sprintf("Policy groups > %s", name),
	})
}
