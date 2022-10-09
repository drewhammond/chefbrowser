package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (s *Service) getEnvironments(c *gin.Context) {
	s.log.Debug("getting all environments from chef server")
	environments, err := s.chef.GetEnvironments(c.Request.Context())
	if err != nil {
		s.log.Error("failed to fetch environments from server", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse("failed to fetch environments from chef server"))
		return
	}

	c.JSON(http.StatusOK, environments)
}

func (s *Service) getEnvironment(c *gin.Context) {
	name := c.Param("name")
	environment, err := s.chef.GetEnvironment(c.Request.Context(), name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("failed to fetch environment from chef server"))
		return
	}
	if environment != nil {
		c.JSON(http.StatusOK, environment)
		return
	}

	c.JSON(http.StatusNotFound, environment)
}
