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

type SearchRecipesRequest struct {
	Query string `json:"query"`
}

func (s *GospigaService) SearchRecipes(c *gin.Context) {
	var req SearchRecipesRequest
	err := c.BindJSON(&req)
	if err != nil {
		c.Error(err)
	}

	recipes, err := s.app.SearchRecipes(c.Copy().Request.Context(), req.Query)
	if err != nil {
		c.Error(err)
	}

	c.JSON(200, recipes)
}
