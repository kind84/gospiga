package grpc

type App interface {
	SearchRecipes(string) ([]string, error)
	AllRecipeTags() ([]string, error)
}
