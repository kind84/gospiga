package api

import (
	"github.com/gin-gonic/gin"
)

func (s *GospigaService) AllTagsImages(c *gin.Context) {
	tags, err := s.app.AllTagsImages(c.Copy().Request.Context())
	if err != nil {
		c.Error(err)
	}
	c.JSON(200, gin.H{"tags": tags})
}
