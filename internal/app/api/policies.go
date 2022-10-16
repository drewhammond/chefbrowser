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

func (s *Service) getPolicyRevision(c *gin.Context) {
	name := c.Param("name")
	revision := c.Param("revision")
	policyRevision, err := s.chef.GetPolicyRevision(c.Request.Context(), name, revision)
	if err != nil {
		s.log.Error("failed to fetch policy revision from server", zap.Error(err))
		c.JSON(http.StatusNotFound, "failed to fetch policy revision from server")
		return
	}
	c.JSON(http.StatusOK, SuccessResponse(policyRevision))
}

func (s *Service) getPolicyGroups(c *gin.Context) {
	policyGroups, err := s.chef.GetPolicyGroups(c.Request.Context())
	if err != nil {
		s.log.Error("failed to fetch policy groups from server", zap.Error(err))
		c.JSON(http.StatusInternalServerError, "failed to fetch policy groups from server")
		return
	}
	c.JSON(http.StatusOK, SuccessResponse(policyGroups))
}

func (s *Service) getPolicyGroup(c *gin.Context) {
	name := c.Param("name")
	policyGroup, err := s.chef.GetPolicyGroup(c.Request.Context(), name)
	if err != nil {
		s.log.Error("failed to fetch policy group from server", zap.Error(err))
		c.JSON(http.StatusNotFound, "failed to fetch policy group from server")
		return
	}
	c.JSON(http.StatusOK, SuccessResponse(policyGroup))
}
