package api

import (
	"net/http"
	"strconv"

	"github.com/drewhammond/chefbrowser/internal/chef"
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
	var nodes *chef.NodeList
	var err error
	if q := c.QueryParam("q"); q != "" {
		limit, _ := strconv.Atoi(c.QueryParam("limit"))
		s.log.Debug("searching nodes on chef server", zap.String("query", q), zap.Int("limit", limit))
		nodes, err = s.chef.SearchNodes(c.Request().Context(), q, limit)
	} else {
		s.log.Debug("getting all nodes from chef server")
		nodes, err = s.chef.GetNodes(c.Request().Context())
	}
	if err != nil {
		s.log.Error("failed to fetch nodes", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, ErrorResponse("failed to fetch nodes"))
	}
	return c.JSON(http.StatusOK, nodes)
}
