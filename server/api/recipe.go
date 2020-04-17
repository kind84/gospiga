package api

import (
	"github.com/gin-gonic/gin"
)

// NewRecipeRequest.
type NewRecipeRequest struct {
	Message  string `json:"message"`
	EntityID string `json:"entity_id"`
}

// NewRecipe listens for new recipe IDs.
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

// UpdatedRecipeRequest.
type UpdatedRecipeRequest struct {
	Message  string `json:"message"`
	EntityID string `json:"entity_id"`
}

// UpdatedRecipe listens for IDs of recipes that have been updated.
func (s *GospigaService) UpdatedRecipe(c *gin.Context) {
	var req UpdatedRecipeRequest
	err := c.BindJSON(&req)
	if err != nil {
		c.Error(err)
	}

	err = s.app.UpdatedRecipe(c.Copy().Request.Context(), req.EntityID)
	if err != nil {
		c.Error(err)
	}
}

// DeletedRecipeRequest.
type DeletedRecipeRequest struct {
	Message  string `json:"message"`
	EntityID string `json:"entity_id"`
}

// DeletedRecipe listens for IDs of recipes that have been deleted.
func (s *GospigaService) DeletedRecipe(c *gin.Context) {
	var req DeletedRecipeRequest
	err := c.BindJSON(&req)
	if err != nil {
		c.Error(err)
	}

	err = s.app.DeletedRecipe(c.Copy().Request.Context(), req.EntityID)
	if err != nil {
		c.Error(err)
	}
}

// LoadRecipes initializes the platform loading all the recipes. It is safe to
// be called multiple times.
func (s *GospigaService) LoadRecipes(c *gin.Context) {
	err := s.app.LoadRecipes(c.Copy().Request.Context())
	if err != nil {
		c.Error(err)
	}
}
