package domain

import (
	"context"
)

type Service struct {
	db DB
}

func NewService(db DB) *Service {
	return &Service{db}
}

func (s *Service) SaveRecipe(ctx context.Context, recipe *Recipe) error {
	return s.db.SaveRecipe(ctx, recipe)
}

func (s *Service) UpdateRecipe(ctx context.Context, recipe *Recipe) error {
	return s.db.UpdateRecipe(ctx, recipe)
}

func (s *Service) GetRecipeByID(ctx context.Context, id string) (*Recipe, error) {
	return s.db.GetRecipeByID(ctx, id)
}

func (s *Service) GetRecipesByIDs(ctx context.Context, ids []string) ([]*Recipe, error) {
	return s.db.GetRecipesByUIDs(ctx, ids)
}

func (s *Service) IDSaved(ctx context.Context, id string) (bool, error) {
	return s.db.IDSaved(ctx, id)
}
