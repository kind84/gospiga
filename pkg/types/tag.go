package types

type Tag struct {
	TagName string    `json:"tagName,omitempty"`
	Recipes []*Recipe `json:"recipes,omitempty"`
}
