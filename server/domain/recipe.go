package domain

type Recipe struct {
	ID          string           `json:"uid,omitempty"`
	ExternalID  string           `json:"id,omitempty"`
	Title       string           `json:"title,omitempty"`
	Subtitle    string           `json:"subtitle,omitempty"`
	MainImage   Image            `json:"mainImage,omitempty"`
	Likes       int              `json:"likes,omitempty"`
	Difficulty  RecipeDifficulty `json:"difficulty,omitempty"`
	Cost        RecipeCost       `json:"cost,omitempty"`
	PrepTime    int              `json:"prepTime,omitempty"`
	CookTime    int              `json:"cookTime,omitempty"`
	Servings    int              `json:"servings,omitempty"`
	ExtraNotes  string           `json:"extraNotes,omitempty"`
	Description string           `json:"description,omitempty"`
	Ingredients []*Ingredient    `json:"ingredients,omitempty"`
	Steps       []*Step          `json:"steps,omitempty"`
	Conclusion  string           `json:"conclusion,omitempty"`
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
	Name          string      `json:"name,omitempty"`
	Quantity      interface{} `json:"quantity,omitempty"`
	UnitOfMeasure string      `json:"unitOfMeasure,omitempty"`
}

type Step struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Image       Image  `json:"image,omitempty"`
}

type Image struct {
	Url string `json:"url,omitempty"`
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
