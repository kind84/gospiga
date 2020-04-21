package types

type Recipe struct {
	ID          string           `json:"uid,omitempty"`
	ExternalID  string           `json:"id,omitempty"`
	Title       string           `json:"title,omitempty"`
	Subtitle    string           `json:"subtitle,omitempty"`
	MainImage   *Image           `json:"mainImage,omitempty"`
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
	Tags        string           `json:"tags,omitempty"`
	Conclusion  string           `json:"conclusion,omitempty"`
	Slug        string           `json:"slug,omitempty"`
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
	Heading string `json:"heading,omitempty"`
	Body    string `json:"body,omitempty"`
	Image   *Image `json:"image,omitempty"`
}

type Image struct {
	URL string `json:"url,omitempty"`
}
