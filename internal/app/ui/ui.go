package ui

import (
	"embed"
	"fmt"
	"github.com/drewhammond/chefbrowser/config"
	"github.com/drewhammond/chefbrowser/internal/chef"
	"github.com/drewhammond/chefbrowser/internal/common/logging"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"html/template"
	"io/fs"
	"net/http"
	"path"
)

//go:embed  all:dist/* templates/* templates/*
var f embed.FS

type Service struct {
	log    *logging.Logger
	config *config.Config
	engine *gin.Engine
	chef   *chef.Service
}

func New(config *config.Config, engine *gin.Engine, chef *chef.Service, logger *logging.Logger) *Service {
	s := Service{
		config: config,
		log:    logger,
		engine: engine,
		chef:   chef,
	}
	return &s
}

func (s *Service) StartServer() {
	s.log.Info(fmt.Sprintf("starting UI server on %s", s.config.App.ListenAddr))
}

func serveFromFS(path string, sub fs.FS) func(c *gin.Context) {
	return func(c *gin.Context) {
		c.FileFromFS(path, http.FS(sub))
	}
}

func (s *Service) RegisterRoutes() {
	s.log.Info(fmt.Sprintf("registering UI routes %s", s.config.App.ListenAddr))

	s.engine.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "/ui")
	})

	//s.engine.NoRoute(func(c *gin.Context) {
	//	sub, _ := fs.Sub(f, "dist")
	//	c.FileFromFS("/", http.FS(sub))
	//})

	uiRouter := s.engine.Group("/ui/_next")
	{
		sub, _ := fs.Sub(f, "dist/_next")
		uiRouter.StaticFS("/", http.FS(sub))
	}

	r2 := s.engine.Group("/ui")
	{
		sub, _ := fs.Sub(f, "dist")

		r2.GET("/", func(c *gin.Context) {
			c.Redirect(http.StatusTemporaryRedirect, "/ui/nodes")
		})

		r2.GET("/nodes", serveFromFS("/nodes.html", sub))
		r2.GET("/environments", serveFromFS("/environments.html", sub))
		r2.GET("/roles", serveFromFS("/roles.html", sub))
		r2.GET("/cookbooks", serveFromFS("/cookbooks.html", sub))
		r2.GET("/groups", serveFromFS("/groups.html", sub))

		fs2 := r2.Group("/node")
		{
			fs2.GET("/*filepath", func(c *gin.Context) {
				c.FileFromFS(path.Join("/node/[id].html"), http.FS(sub))
			})
		}

		grp := r2.Group("/group")
		{
			grp.GET("/*filepath", func(c *gin.Context) {
				c.FileFromFS(path.Join("/group/[id].html"), http.FS(sub))
			})
		}

		//fs2 := r2.Group("/node")
		//{
		//	fs2.GET("/*filepath", func(c *gin.Context) {
		//		c.FileFromFS(path.Join("/node/[id].html"), http.FS(sub))
		//	})
		//}

		//r2.GET("/*filepath", func(c *gin.Context) {
		//	c.FileFromFS(path.Join("/"), http.FS(sub))
		//})
		//uiRouter.StaticFileFS("/", http.FS(sub))
	}

	//r3 := s.engine.Group("/ui/node")
	//{
	//	sub, _ := fs.Sub(f, "dist")
	//	r3.GET("/*filepath", func(c *gin.Context) {
	//		c.FileFromFS(path.Join("/node/[id].html"), http.FS(sub))
	//	})
	//	//uiRouter.StaticFileFS("/", http.FS(sub))
	//}
	//
	//s.engine.NoRoute(func(c *gin.Context) {
	//	sub, _ := fs.Sub(f, "dist")
	//	c.FileFromFS("/ui/index.html", http.FS(sub))
	//})

	////s.engine.StaticFileFS("/foos", "bar", http.FS(sub))
	////s.engine.StaticFileFS("/foo/:name", "foo", http.FS(sub))
	//s.engine.StaticFS("/nodes", http.FS(sub))
	//
	////s.engine.GET("/nodes", s.getNodes)
	//s.engine.GET("/node/:name", s.getNode)

	//router := s.engine.Group("/ui")
	//{
	//	// nodes
	//	//router.GET("/nodes", s.getNodes)
	//
	//	sub, _ := fs.Sub(f, "dist")
	//	router.StaticFS("/", http.FS(sub))
	//}
}

func (s *Service) RegisterRoutesWithTemplates() {

	//templ := template.Must(
	//	template.New("").ParseFS(f, "templates/*.tmpl.html", "templates/partials/*.html"))

	//files := []string{
	//	".templates/base.tmpl",
	//	".templates/partials/nav.tmpl",
	//	".templates/pages/home.tmpl",
	//}

	//s.engine.LoadHTMLGlob("templates/**/*.html")

	ts, err := template.Must(template.New(""), nil).ParseFS(f, "templates/*.html", "templates/**/*.html")
	if err != nil {
		s.log.Error("failed to parse templates")
	}

	fmt.Println(ts.DefinedTemplates())

	fmt.Println(ts)

	//s.engine.SetHTMLTemplate(ts)
	ui := s.engine.Group("/ui")
	{

		// example: /public/assets/images/example.png
		ui.StaticFS("/public", http.FS(f))

		ui.GET("/", func(c *gin.Context) {

			ts.ExecuteTemplate(c.Writer, "base", gin.H{})

			//c.HTML(http.StatusOK, "base.tmpl.html", gin.H{})

			//c.HTML(http.StatusOK, "pages/nodes.tmpl.html", gin.H{
			//	"title": "Home Page",
			////})
			//ts.ExecuteTemplate(c.Writer, "nodes.tmpl.html", nil)
			////c.HTML(http.StatusOK, "nodes.tmpl.html", gin.H{
			////	"title": "index",
			////})
		})

		//ui.GET("/environments", s.getEnvironments)
		//ui.GET("/roles", s.getRoles)
		//ui.GET("/databags", s.getDatabags)
		//ui.GET("/cookbooks", s.getCookbooks)
		//
		//ui.GET("/nodes", s.getNodes)
		//ui.GET("/node/:name", s.getNode)
		//ui.GET("/environment/:name", s.getEnvironment)
	}

}

func (s *Service) getNodes(c *gin.Context) {

	c.HTML(http.StatusOK, "base.tmpl.html", gin.H{})
	//
	//s.log.Debug("getting all nodes from chef server")
	//nodes, err := s.chef.GetNodes(c.Request.Context())
	//fmt.Println(nodes)
	//if err != nil {
	//	s.log.Error("failed to load chef node", zap.Error(err))
	//	// todo: return error page
	//}
	//c.HTML(http.StatusOK, "nodes.tmpl", gin.H{})
}

func (s *Service) getNode(c *gin.Context) {
	name := c.Param("name")
	node, err := s.chef.GetNode(c.Request.Context(), name)
	if err != nil {
		s.log.Error("failed to load chef node", zap.Error(err))
		// todo: return error page
	}
	c.HTML(http.StatusOK, "show.tmpl", gin.H{
		"title": node.Name,
	})
}

func (s *Service) getEnvironments(c *gin.Context) {

	c.HTML(http.StatusOK, "nodes.tmpl.html", gin.H{})
	//environments, err := s.chef.GetEnvironments(c.Request.Context())
	//if err != nil {
	//	s.log.Error("failed to load chef node", zap.Error(err))
	//	// todo: return error page
	//}
	//
	//c.HTML(http.StatusOK, "environments.tmpl", gin.H{
	//	"title": environments,
	//})
}

func (s *Service) getRoles(c *gin.Context) {
	roles, err := s.chef.GetRoles(c.Request.Context())
	if err != nil {
		s.log.Error("failed to load chef node", zap.Error(err))
		// todo: return error page
	}

	c.HTML(http.StatusOK, "roles.tmpl", gin.H{
		"title": roles,
	})
}

func (s *Service) getCookbooks(c *gin.Context) {
	cookbooks, err := s.chef.GetCookbooks(c.Request.Context())
	if err != nil {
		s.log.Error("failed to load chef node", zap.Error(err))
		// todo: return error page
	}

	c.HTML(http.StatusOK, "cookbooks.tmpl", gin.H{
		"title": cookbooks,
	})
}

func (s *Service) getDatabags(c *gin.Context) {

	c.HTML(http.StatusOK, "databags.tmpl", gin.H{
		"title": "foo",
	})
}

func (s *Service) getEnvironment(c *gin.Context) {
	name := c.Param("name")
	environment, err := s.chef.GetEnvironment(c.Request.Context(), name)
	if err != nil {
		s.log.Error("failed to load chef node", zap.Error(err))
		// todo: return error page
	}

	c.HTML(http.StatusOK, "environment_show.tmpl", gin.H{
		"title": environment.Name,
	})
}
