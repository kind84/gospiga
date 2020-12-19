package domain

import (
	"context"

	"gospiga/pkg/types"
)

// service implements the domain service interface.
type service struct {
	db DB
}

// NewService constructor.
func NewService(db DB) *service {
	return &service{db}
}

func (s *service) SaveRecipe(ctx context.Context, recipe *Recipe) error {
	return s.db.SaveRecipe(ctx, recipe)
}

func (s *service) UpdateRecipe(ctx context.Context, recipe *Recipe) (string, error) {
	return s.db.UpdateRecipe(ctx, recipe)
}

func (s *service) DeleteRecipe(ctx context.Context, recipeID string) error {
	return s.db.DeleteRecipe(ctx, recipeID)
}

func (s *service) GetRecipeByID(ctx context.Context, id string) (*Recipe, error) {
	return s.db.GetRecipeByID(ctx, id)
}

func (s *service) GetRecipesByIDs(ctx context.Context, ids []string) ([]*Recipe, error) {
	return s.db.GetRecipesByUIDs(ctx, ids)
}

func (s *service) IDSaved(ctx context.Context, id string) (bool, error) {
	return s.db.IDSaved(ctx, id)
}

func (s *service) SearchRecipes(ctx context.Context, args *types.SearchRecipesArgs) ([]*Recipe, error) {
	return s.db.SearchRecipes(ctx, args)
}
