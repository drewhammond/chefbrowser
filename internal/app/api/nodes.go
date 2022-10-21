package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func (s *Service) getNode(c echo.Context) error {
	name := c.Param("name")
	s.log.Debug("getting node from chef server")
	node, err := s.chef.GetNode(c.Request().Context(), name)
	if err != nil {
		s.log.Error("failed to fetch node from server", zap.Error(err))
	}
	return c.JSON(http.StatusOK, node)
}

func (s *Service) getNodes(c echo.Context) error {
	s.log.Debug("getting all nodes from chef server")
	nodes, err := s.chef.GetNodes(c.Request().Context())
	if err != nil {
		s.log.Error("failed to fetch nodes", zap.Error(err))
	}
	return c.JSON(http.StatusOK, nodes)
}
