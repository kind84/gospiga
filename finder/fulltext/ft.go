package fulltext

import (
	"github.com/kind84/gospiga/finder/domain"
)

type FT interface {
	IndexRecipe(*domain.Recipe) error
	DeleteRecipe(string) error
	SearchRecipes(string) ([]string, error)
}
