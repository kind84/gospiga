package domain

type Recipe struct {
	ID        string
	Title     string
	Subtitle  string
	MainImage struct {
		Url string
	}
	Likes       int
	Difficulty  RecipeDifficulty
	Cost        RecipeCost
	PrepTime    int
	CookTime    int
	Servings    int
	ExtraNotes  string
	Description string
	Ingredients []*Ingredient
	Steps       []*Step
	Conclusion  string
}

type RecipeDifficulty string

const (
	DifficultyEasy = "Bassa"
	DifficultyMid  = "Media"
	DifficultyHard = "Alta"
)

type RecipeCost string

const (
	CostLow  = "Basso"
	CostMid  = "Medio"
	CostHigh = "Alto"
)

type Ingredient struct {
	Name          string
	Quantity      interface{}
	UnitOfMeasure string
}

type Step struct {
	Title       string
	Description string
	Image       struct {
		Url string
	}
}

func NewRecipe(id string, title string) (*Recipe, error) {
	r := &Recipe{
		ID:    id,
		Title: title,
	}
	return r, nil
}

func (r *Recipe) Hello() string {
	return r.Title
}
