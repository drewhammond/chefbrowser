package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (s *Service) getGroups(c *gin.Context) {
	groups, err := s.chef.GetGroups(c.Request.Context())
	if err != nil {
		s.log.Error("failed to fetch groups from server", zap.Error(err))
		c.JSON(http.StatusInternalServerError, "failed to fetch groups from server")
		return
	}
	c.JSON(http.StatusOK, SuccessResponse(groups))
}

func (s *Service) getGroup(c *gin.Context) {
	name := c.Param("name")
	group, err := s.chef.GetGroup(c.Request.Context(), name)
	if err != nil {
		s.log.Error("failed to fetch group from server", zap.Error(err))
		c.JSON(http.StatusNotFound, "failed to fetch group from server")
		return
	}
	c.JSON(http.StatusOK, group)
}
