package api

import (
	"gospiga/finder/fulltext"
)

type App interface {
	SearchRecipes(string) ([]*fulltext.Recipe, error)
	SearchByTag([]string) ([]*fulltext.Recipe, error)
	AllRecipeTags() ([]string, error)
}
