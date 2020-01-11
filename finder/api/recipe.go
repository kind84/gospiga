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

	ids, err := s.app.Search(c.Copy().Request.Context(), req.Query)
	if err != nil {
		c.Error(err)
	}
	c.JSON(200, gin.H{
		"ids": ids,
	})
}
