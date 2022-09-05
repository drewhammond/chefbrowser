package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (s *Service) getDatabags(c *gin.Context) {
	s.log.Debug("getting all databags from chef server")
	databags, err := s.chef.GetDatabags(c.Request.Context())
	if err != nil {
		s.log.Error("failed to fetch databags from server", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse("failed to fetch databags from server"))
		return
	}
	c.JSON(http.StatusOK, databags)
}

func (s *Service) getDatabagItems(c *gin.Context) {
	name := c.Param("name")
	databag, err := s.chef.GetDatabagItems(c.Request.Context(), name)
	if err != nil {
		s.log.Error("failed to fetch databag from server", zap.Error(err))
		c.JSON(http.StatusNotFound, ErrorResponse("failed to fetch databag from server"))
		return
	}
	c.JSON(http.StatusOK, databag)
}

func (s *Service) getDatabagItemContent(c *gin.Context) {
	name := c.Param("name")
	item := c.Param("item")
	content, err := s.chef.GetDatabagItemContent(c.Request.Context(), name, item)
	if err != nil {
		s.log.Error("failed to fetch databag contents from server", zap.Error(err))
		c.JSON(http.StatusNotFound, ErrorResponse("failed to fetch databag contents from server"))
		return
	}
	c.JSON(http.StatusOK, content)
}
