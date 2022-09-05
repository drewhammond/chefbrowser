package api

import (
	"fmt"
	"net/http"

	"github.com/drewhammond/chefbrowser/config"
	"github.com/drewhammond/chefbrowser/internal/chef"
	"github.com/drewhammond/chefbrowser/internal/common/logging"
	"github.com/gin-gonic/gin"
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
	s.log.Info(fmt.Sprintf("registering API routes %s", s.config.App.ListenAddr))

	router := s.engine.Group("/api")
	{
		router.Use(corsMiddleware())
		// nodes
		router.GET("/nodes", s.getNodes)
		router.GET("/node/:name", s.getNode)

		// environments
		router.GET("/environments", s.getEnvironments)
		router.GET("/environment/:name", s.getEnvironment)

		// roles
		router.GET("/roles", s.getRoles)
		router.GET("/role/:name", s.getRole)

		// cookbooks
		router.GET("/cookbooks", s.getCookbooks)
		router.GET("/cookbook/:name", s.getCookbook)
		router.GET("/cookbook/:name/:version", s.getCookbookVersion)

		// groups
		router.GET("/groups", s.getGroups)
		router.GET("/groups/:name", s.getGroup)

		// misc
		router.GET("/health", getHealth)
	}
}

type HealthResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func getHealth(c *gin.Context) {
	c.JSON(http.StatusOK, &HealthResponse{Success: true, Message: "ready"})
}

// FIXME: we shouldn't use this by default; this is just to help with my local testing
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "*")
		c.Next()
	}
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
