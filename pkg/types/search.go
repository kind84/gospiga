package types

type SearchRecipesArgs struct {
	First       *int
	After       *string
	Tags        []string
	Ingredients []string
	Query       *string
	IDs         []string
}
