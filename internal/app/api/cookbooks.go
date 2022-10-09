package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (s *Service) getCookbooks(c *gin.Context) {
	s.log.Debug("getting all cookbooks from chef server")
	cookbooks, err := s.chef.GetCookbooks(c.Request.Context())
	if err != nil {
		s.log.Error("failed to fetch cookbooks from server", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse("failed to fetch cookbooks from server"))
		return
	}
	c.JSON(http.StatusOK, cookbooks)
}

func (s *Service) getCookbook(c *gin.Context) {
	name := c.Param("name")
	cookbook, err := s.chef.GetCookbook(c.Request.Context(), name)
	if err != nil {
		s.log.Error("failed to fetch cookbook from server", zap.Error(err))
		c.JSON(http.StatusNotFound, ErrorResponse("failed to fetch cookbook from server"))
		return
	}
	c.JSON(http.StatusOK, cookbook)
}

func (s *Service) getCookbookVersion(c *gin.Context) {
	name := c.Param("name")
	version := c.Param("version")
	cookbook, err := s.chef.GetCookbookVersion(c.Request.Context(), name, version)
	if err != nil {
		s.log.Error("failed to fetch cookbook from server", zap.Error(err))
		c.JSON(http.StatusNotFound, ErrorResponse("failed to fetch cookbook version from server"))
		return
	}
	c.JSON(http.StatusOK, cookbook)
}

func (s *Service) getCookbookVersions(c *gin.Context) {
	name := c.Param("name")

	resp, err := s.chef.GetClient().Cookbooks.GetAvailableVersions(name, "0")
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse("failed to fetch cookbook versions"))
		return
	}

	var versions []string
	for _, i := range resp {
		for _, j := range i.Versions {
			versions = append(versions, j.Version)
		}
	}

	c.JSON(http.StatusOK, versions)
}
