package api

import (
	"github.com/gin-gonic/gin"
)

func (s *GospigaService) NewRecipe(c *gin.Context) {
	err := s.app.NewRecipe(c.Copy().Request.Context(), c.Request.Body)
	if err != nil {
		c.Error(err)
	}
}
