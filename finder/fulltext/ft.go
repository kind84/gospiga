package fulltext

import (
	"gospiga/finder/domain"
)

type FT interface {
	IndexRecipe(*domain.Recipe) error
	DeleteRecipe(string) error
	SearchRecipes(string) ([]string, error)
}
