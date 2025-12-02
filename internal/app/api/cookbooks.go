package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func (s *Service) getCookbooks(c echo.Context) error {
	s.log.Debug("getting all cookbooks from chef server")
	cookbooks, err := s.chef.GetCookbooks(c.Request().Context())
	if err != nil {
		s.log.Error("failed to fetch cookbooks from server", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, ErrorResponse("failed to fetch cookbooks from server"))
	}
	return c.JSON(http.StatusOK, cookbooks)
}

func (s *Service) getCookbook(c echo.Context) error {
	name := c.Param("name")
	cookbook, err := s.chef.GetCookbook(c.Request().Context(), name)
	if err != nil {
		s.log.Error("failed to fetch cookbook from server", zap.Error(err))
		return c.JSON(http.StatusNotFound, ErrorResponse("failed to fetch cookbook from server"))
	}
	return c.JSON(http.StatusOK, cookbook)
}

func (s *Service) getCookbookVersion(c echo.Context) error {
	name := c.Param("name")
	version := c.Param("version")
	cookbook, err := s.chef.GetCookbookVersion(c.Request().Context(), name, version)
	if err != nil {
		s.log.Error("failed to fetch cookbook from server", zap.Error(err))
		return c.JSON(http.StatusNotFound, ErrorResponse("failed to fetch cookbook version from server"))
	}
	return c.JSON(http.StatusOK, cookbook)
}

func (s *Service) getCookbookVersions(c echo.Context) error {
	name := c.Param("name")

	versions, err := s.chef.GetCookbookVersions(c.Request().Context(), name)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse("failed to fetch cookbook versions"))
	}

	return c.JSON(http.StatusOK, versions)
}
