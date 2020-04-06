package fulltext

import (
	"encoding/json"
	"strconv"
)

type Recipe struct {
	ID          string `json:"id,omitempty"`
	ExternalID  string `json:"xid,omitempty"`
	Title       string `json:"title,omitempty"`
	Subtitle    string `json:"subtitle,omitempty"`
	MainImage   *Image `json:"mainImage,omitempty"`
	PrepTime    int    `json:"prepTime,omitempty"`
	CookTime    int    `json:"cookTime,omitempty"`
	Description string `json:"description,omitempty"`
	Ingredients string `json:"ingredients,omitempty"`
	Steps       string `json:"steps,omitempty"`
	Tags        string `json:"tags,omitempty"`
	Conclusion  string `json:"conclusion,omitempty"`
	// Likes        int              `json:"likes,omitempty"`
	// Difficulty   RecipeDifficulty `json:"difficulty,omitempty"`
	// Cost         RecipeCost       `json:"cost,omitempty"`
	// PrepTime     int              `json:"prepTime,omitempty"`
	// CookTime     int              `json:"cookTime,omitempty"`
	// Servings     int              `json:"servings,omitempty"`
	// ExtraNotes   string           `json:"extraNotes,omitempty"`
}

type Image struct {
	URL string `json:"url"`
}

func (r *Recipe) UnmarshalJSON(b []byte) error {
	type recipe struct {
		ID          string `json:"id,omitempty"`
		ExternalID  string `json:"xid,omitempty"`
		Title       string `json:"title,omitempty"`
		Subtitle    string `json:"subtitle,omitempty"`
		MainImage   string `json:"mainImage,omitempty"`
		PrepTime    string `json:"prepTime,omitempty"`
		CookTime    string `json:"cookTime,omitempty"`
		Description string `json:"description,omitempty"`
		Ingredients string `json:"ingredients,omitempty"`
		Steps       string `json:"steps,omitempty"`
		Tags        string `json:"tags,omitempty"`
		Conclusion  string `json:"conclusion,omitempty"`
		// Likes        int              `json:"likes,omitempty"`
		// Difficulty   RecipeDifficulty `json:"difficulty,omitempty"`
		// Cost         RecipeCost       `json:"cost,omitempty"`
		// PrepTime     int              `json:"prepTime,omitempty"`
		// CookTime     int              `json:"cookTime,omitempty"`
		// Servings     int              `json:"servings,omitempty"`
		// ExtraNotes   string           `json:"extraNotes,omitempty"`
	}

	var rcp recipe
	err := json.Unmarshal(b, &rcp)
	if err != nil {
		return err
	}

	ptime, err := strconv.Atoi(rcp.PrepTime)
	if err != nil {
		return err
	}
	ctime, err := strconv.Atoi(rcp.CookTime)
	if err != nil {
		return err
	}

	r.ID = rcp.ID
	r.ExternalID = rcp.ExternalID
	r.Title = rcp.Title
	r.Subtitle = rcp.Subtitle
	r.MainImage = &Image{URL: rcp.MainImage}
	r.PrepTime = ptime
	r.CookTime = ctime
	r.Description = rcp.Description
	r.Ingredients = rcp.Ingredients
	r.Steps = rcp.Steps
	r.Tags = rcp.Tags
	r.Conclusion = rcp.Conclusion

	return nil
}
