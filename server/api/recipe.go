package api

import (
	"github.com/gin-gonic/gin"
)

type NewRecipeRequest struct {
	Message  string `json:"message"`
	EntityID string `json:"entity_id"`
}

func (s *GospigaService) NewRecipe(c *gin.Context) {
	var req NewRecipeRequest
	err := c.BindJSON(&req)
	if err != nil {
		c.Error(err)
	}

	err = s.app.NewRecipe(c.Copy().Request.Context(), req.EntityID)
	if err != nil {
		c.Error(err)
	}
}
