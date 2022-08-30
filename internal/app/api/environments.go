package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
)

func (s *Service) getEnvironments(c *gin.Context) {
	s.log.Debug("getting all environments from chef server")
	environments, err := s.chef.GetEnvironments(c.Request.Context())
	if err != nil {
		s.log.Error("failed to fetch environments from server", zap.Error(err))
	}

	c.JSON(http.StatusOK, environments)
}

type getEnvironmentsResponse struct {
	Name string `json:"name"`
}

func (s *Service) getEnvironment(c *gin.Context) {
	name := c.Param("name")
	s.log.Debug(fmt.Sprintf("getting environment %s from chef server", name))
	environment, err := s.chef.GetEnvironment(c.Request.Context(), name)
	if err != nil {
		s.log.Error(fmt.Sprintf("failed to fetch environment %s from server", name), zap.Error(err))
	}
	if environment != nil {
		c.JSON(http.StatusOK, environment)
		return
	}

	c.JSON(http.StatusNotFound, environment)
}
