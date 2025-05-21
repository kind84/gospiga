package api

import (
	"context"
	"gospiga/finder/fulltext"
)

type App interface {
	SearchRecipes(string) ([]*fulltext.Recipe, error)
	SearchByTag([]string) ([]*fulltext.Recipe, error)
	AllRecipeTags(context.Context) ([]string, error)
}
