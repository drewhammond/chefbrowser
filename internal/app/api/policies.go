package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (s *Service) getPolicies(c *gin.Context) {
	policies, err := s.chef.GetPolicies(c.Request.Context())
	if err != nil {
		s.log.Error("failed to fetch policies from server", zap.Error(err))
		c.JSON(http.StatusInternalServerError, "failed to fetch policies from server")
		return
	}
	c.JSON(http.StatusOK, SuccessResponse(policies))
}

func (s *Service) getPolicy(c *gin.Context) {
	name := c.Param("name")
	policies, err := s.chef.GetPolicy(c.Request.Context(), name)
	if err != nil {
		s.log.Error("failed to fetch policy from server", zap.Error(err))
		c.JSON(http.StatusNotFound, "failed to fetch policy from server")
		return
	}
	c.JSON(http.StatusOK, SuccessResponse(policies))
}
