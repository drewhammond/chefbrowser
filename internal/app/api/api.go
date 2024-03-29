package api

import (
	"net/http"

	"github.com/drewhammond/chefbrowser/config"
	"github.com/drewhammond/chefbrowser/internal/chef"
	"github.com/drewhammond/chefbrowser/internal/common/logging"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var basePath = ""

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
	basePath = config.Server.BasePath
	return &s
}

func (s *Service) RegisterRoutes() {
	s.log.Info("registering API routes")

	router := s.engine.Group(urlWithBasePath("/api"))
	{
		router.Use(middleware.CORS())
		// nodes
		router.GET("/nodes", s.getNodes)
		router.GET("/nodes/:name", s.getNode)

		// environments
		router.GET("/environments", s.getEnvironments)
		router.GET("/environments/:name", s.getEnvironment)

		// roles
		router.GET("/roles", s.getRoles)
		router.GET("/roles/:name", s.getRole)

		// cookbooks
		router.GET("/cookbooks", s.getCookbooks)
		router.GET("/cookbooks/:name", s.getCookbook)
		router.GET("/cookbooks/:name/versions", s.getCookbookVersions)
		router.GET("/cookbooks/:name/:version", s.getCookbookVersion)

		// groups
		router.GET("/groups", s.getGroups)
		router.GET("/groups/:name", s.getGroup)

		// databags
		router.GET("/databags", s.getDatabags)
		router.GET("/databags/:name", s.getDatabagItems)
		router.GET("/databags/:name/:item", s.getDatabagItemContent)

		// policies
		router.GET("/policies", s.getPolicies)
		router.GET("/policies/:name", s.getPolicy)
		router.GET("/policies/:name/:revision", s.getPolicyRevision)
		router.GET("/policy-groups", s.getPolicyGroups)
		router.GET("/policy-groups/:name", s.getPolicyGroup)

		// misc
		router.GET("/health", getHealth)
	}
}

func urlWithBasePath(path string) string {
	return basePath + path
}

type HealthResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func getHealth(c echo.Context) error {
	return c.JSON(http.StatusOK, &HealthResponse{Success: true, Message: "ready"})
}

type errorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type successResponse struct {
	Success bool        `json:"success"`
	Results interface{} `json:"message"`
}

func ErrorResponse(message string) errorResponse {
	return errorResponse{
		Success: false,
		Message: message,
	}
}

func SuccessResponse(body interface{}) successResponse {
	return successResponse{
		Success: true,
		Results: body,
	}
}
