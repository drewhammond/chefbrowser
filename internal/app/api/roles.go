package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func (s *Service) getRoles(c echo.Context) error {
	s.log.Debug("getting all roles from chef server")
	roles, err := s.chef.GetRoles(c.Request().Context())
	if err != nil {
		s.log.Error("failed to fetch roles from server", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, ErrorResponse("failed to fetch roles from server"))
	}
	return c.JSON(http.StatusOK, roles)
}

func (s *Service) getRole(c echo.Context) error {
	name := c.Param("name")
	s.log.Debug("getting role from chef server")
	role, err := s.chef.GetRole(c.Request().Context(), name)
	if err != nil {
		s.log.Error("failed to fetch role from server", zap.Error(err))
		return c.JSON(http.StatusNotFound, ErrorResponse("failed to fetch role from server"))
	}
	return c.JSON(http.StatusOK, role)
}
