package api

import (
	"github.com/kind84/gospiga/finder/fulltext"
)

type App interface {
	SearchRecipes(string) ([]*fulltext.Recipe, error)
	AllRecipeTags() ([]string, error)
}
