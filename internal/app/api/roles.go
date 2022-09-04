package api

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
)

func (s *Service) getRoles(c *gin.Context) {
	s.log.Debug("getting all roles from chef server")
	roles, err := s.chef.GetRoles(c.Request.Context())
	if err != nil {
		s.log.Error("failed to fetch roles from server", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse("failed to fetch roles from server"))
		return
	}
	c.JSON(http.StatusOK, roles)
}

func (s *Service) getRole(c *gin.Context) {
	name := c.Param("name")
	s.log.Debug("getting role from chef server")
	role, err := s.chef.GetRole(c.Request.Context(), name)
	if err != nil {
		s.log.Error("failed to fetch role from server", zap.Error(err))
		c.JSON(http.StatusNotFound, ErrorResponse("failed to fetch role from server"))
		return
	}
	c.JSON(http.StatusOK, role)
}
