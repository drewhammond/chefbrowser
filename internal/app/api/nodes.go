package api

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
)

func (s *Service) getNode(c *gin.Context) {
	name := c.Param("name")
	s.log.Debug("getting node from chef server")
	node, err := s.chef.GetNode(c.Request.Context(), name)
	if err != nil {
		s.log.Error("failed to fetch node from server", zap.Error(err))

	}
	c.JSON(http.StatusOK, node)
}

func (s *Service) getNodes(c *gin.Context) {
	s.log.Debug("getting all nodes from chef server")
	nodes, err := s.chef.GetNodes(c.Request.Context())
	if err != nil {
		s.log.Error("failed to fetch nodes", zap.Error(err))
	}
	c.JSON(http.StatusOK, nodes)
}
