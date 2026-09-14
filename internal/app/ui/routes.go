package ui

import (
	"crypto/tls"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/drewhammond/chefbrowser/config"
	"github.com/drewhammond/chefbrowser/internal/chef"
	"github.com/drewhammond/chefbrowser/internal/common/logging"
	"github.com/drewhammond/chefbrowser/internal/common/version"
	"github.com/drewhammond/chefbrowser/ui"
	"github.com/foolin/goview"
	"github.com/foolin/goview/supports/echoview-v4"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

var (
	viteFS   = echo.MustSubFS(ui.Embedded, "dist")
	basePath = ""
)

func embeddedFH(config goview.Config, tmpl string) (string, error) {
	path := filepath.Join(config.Root, tmpl)
	bytes, err := ui.Embedded.ReadFile(path + config.Extension)
	return string(bytes), err
}

type Service struct {
	log         *logging.Logger
	config      *config.Config
	chef        chef.Interface
	engine      *echo.Echo
	customLinks *CustomLinksCollection
}

type CustomLink struct {
	Title  string `json:"title"`
	Href   string `json:"href"`
	NewTab bool   `json:"new_tab"`
}

type CustomLinksCollection struct {
	Nodes        []CustomLink
	Environments []CustomLink // Unused, but maybe in the future
	Roles        []CustomLink // Unused, but maybe in the future
	DataBags     []CustomLink // Unused, but maybe in the future
}

func New(config *config.Config, engine *echo.Echo, chef chef.Interface, logger *logging.Logger) *Service {
	s := Service{
		config: config,
		chef:   chef,
		log:    logger,
		engine: engine,
	}
	basePath = config.Server.BasePath
	return &s
}

func (s *Service) RegisterRoutes() {
	s.log.Info("registering UI routes")

	templateRoot := "templates"
	disableCache := false
	if s.config.App.AppMode == "development" {
		s.log.Warn("development mode enabled! view cache is disabled and templates are not loaded from embed.FS")
		templateRoot = "ui/templates"
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

	vCfg := ViteConfig{
		Environment: s.config.App.AppMode,
		Base:        urlWithBasePath("/ui"),
	}

	if s.config.App.AppMode == "production" {
		mf, _ := ui.Embedded.ReadFile("dist/manifest.json")
		vCfg.Manifest = mf
	}

	vite, err := NewVite(vCfg)
	if err != nil {
		s.log.Error("failed to set up vite")
	}

	err = s.BuildCustomLinks()
	if err != nil {
		s.log.Error("failed to validate custom links configuration", zap.Error(err))
	}

	viteTags := vite.HTMLTags
	cfg.Funcs["makeRunListURL"] = s.makeRunListURL
	cfg.Funcs["base_path"] = func() string { return basePath }
	cfg.Funcs["app_version"] = func() string { return version.Get().Version }
	cfg.Funcs["vite_assets"] = func() template.HTML {
		return template.HTML(viteTags)
	}
	cfg.Funcs["add"] = func(a, b int) int { return a + b }
	cfg.Funcs["sub"] = func(a, b int) int { return a - b }

	ev := echoview.New(cfg)
	if s.config.App.AppMode == "production" {
		ev.ViewEngine.SetFileHandler(embeddedFH)
	}

	s.engine.Renderer = ev

	s.engine.GET(urlWithBasePath(""), func(c echo.Context) error {
		return c.Redirect(http.StatusFound, urlWithBasePath("/ui/nodes"))
	})

	// Always redirect to base path if somehow bypassed
	s.engine.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusFound, urlWithBasePath("/ui/nodes"))
	})

	s.engine.GET("/robots.txt", ViteHandler(""))

	s.engine.RouteNotFound("/*", func(c echo.Context) error {
		return c.Render(http.StatusNotFound, "errors/404", echo.Map{
			"message": "Invalid route!",
		})
	})

	router := s.engine.Group(vCfg.Base)
	{
		router.GET("/", func(c echo.Context) error {
			return c.Redirect(http.StatusFound, urlWithBasePath("/ui/nodes"))
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

		router.GET("/assets/*", ViteHandler(vCfg.Base), CacheControlMiddleware)
		router.GET("/favicons/*", ViteHandler(vCfg.Base), CacheControlMiddleware)
	}
}

// BuildCustomLinks returns a map of custom links to be displayed in the UI
func (s *Service) BuildCustomLinks() error {
	clc := CustomLinksCollection{}
	nodeLinks := s.config.CustomLinks.Nodes

	// Sort keys for deterministic ordering
	keys := make([]int, 0, len(nodeLinks))
	for key := range nodeLinks {
		keys = append(keys, key)
	}

	sort.Ints(keys)

	for _, key := range keys {
		clc.Nodes = append(clc.Nodes, CustomLink{
			Title:  nodeLinks[key].Title,
			Href:   nodeLinks[key].Href,
			NewTab: nodeLinks[key].NewTab,
		})
	}

	s.customLinks = &clc
	return nil
}

// CacheControlMiddleware adds Cache-Control headers to static assets so that browsers can cache them
// for subsequent requests. Note that this should only be used on unique filenames such as those generated
// by the build process.
func CacheControlMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Add(echo.HeaderCacheControl, "public,max-age=31536000,immutable")
		return next(c)
	}
}

func ViteHandler(prefix string) echo.HandlerFunc {
	fs := http.FS(viteFS)
	h := http.StripPrefix(prefix, http.FileServer(fs))
	return echo.WrapHandler(h)
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
		"active_nav":   "nodes",
		"custom_links": s.customLinks.Nodes,
		"node":         node,
		"title":        node.Name,
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

const maxPageSize = 10000

func (s *Service) getNodes(c echo.Context) error {
	query := c.QueryParam("q")
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if perPage < 0 {
		perPage = 0
	}

	effectivePerPage := perPage
	if effectivePerPage == 0 {
		effectivePerPage = maxPageSize
	}

	start := (page - 1) * effectivePerPage

	var result *chef.NodeListResult
	var err error

	if query != "" {
		searchQuery := query
		if !strings.Contains(query, ":") {
			searchQuery = fuzzifySearchStr(query)
		}
		result, err = s.chef.SearchNodesWithDetails(c.Request().Context(), searchQuery, start, effectivePerPage)
	} else {
		result, err = s.chef.GetNodesWithDetails(c.Request().Context(), start, effectivePerPage)
	}

	if err != nil {
		s.log.Error("failed to fetch nodes", zap.Error(err))
		return c.Render(http.StatusInternalServerError, "errors/500", echo.Map{
			"message": "failed to fetch nodes",
		})
	}

	totalPages := 1
	if perPage > 0 {
		totalPages = (result.Total + perPage - 1) / perPage
	}

	return c.Render(http.StatusOK, "nodes", echo.Map{
		"nodes":          result.Nodes,
		"total":          result.Total,
		"page":           page,
		"per_page":       perPage,
		"total_pages":    totalPages,
		"query":          query,
		"active_nav":     "nodes",
		"search_enabled": true,
		"title":          "All Nodes",
	})
}

// escapeSolrSpecialChars escapes characters that have special meaning in Solr/Lucene query syntax
func escapeSolrSpecialChars(s string) string {
	specialChars := []string{"\\", "+", "-", "&&", "||", "!", "(", ")", "{", "}", "[", "]", "^", "\"", "~", "?", ":", "/"}
	result := s
	for _, char := range specialChars {
		result = strings.ReplaceAll(result, char, "\\"+char)
	}
	return result
}

// fuzzifySearchStr mimics the fuzzy search functionality
// provided by chef https://github.com/chef/chef/blob/main/lib/chef/search/query.rb#L109
func fuzzifySearchStr(s string) string {
	escaped := escapeSolrSpecialChars(s)
	format := []string{
		"tags:*%v*",
		"roles:*%v*",
		"fqdn:*%v*",
		"addresses:*%v*",
		"policy_name:*%v*",
		"policy_group:*%v*",
	}
	var b strings.Builder
	for i, f := range format {
		if i > 0 {
			b.WriteString(" OR ")
		}
		b.WriteString(fmt.Sprintf(f, escaped))
	}
	return b.String()
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

func urlWithBasePath(path string) string {
	return basePath + path
}
