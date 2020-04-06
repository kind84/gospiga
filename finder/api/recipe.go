package api

import (
	"github.com/gin-gonic/gin"
)

type SearchRequest struct {
	Query string `json:"query"`
}

func (s *GospigaService) SearchRecipes(c *gin.Context) {
	var req SearchRequest
	err := c.BindJSON(&req)
	if err != nil {
		c.Error(err)
	}

	recipes, err := s.app.SearchRecipes(req.Query)
	if err != nil {
		c.Error(err)
	}
	c.JSON(200, gin.H{"recipes": recipes})
}

func (s *GospigaService) AllRecipeTags(c *gin.Context) {
	tags, err := s.app.AllRecipeTags()
	if err != nil {
		c.Error(err)
	}
	c.JSON(200, gin.H{"tags": tags})
}
