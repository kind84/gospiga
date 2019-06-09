package models

import "time"

type Recipe struct {
	UID         string    `json:"uid,omitempty"`
	Title       string    `json:"title,omitempty"`
	Subtitle    string    `json:"subtitle,omitempty"`
	Description string    `json:"description,omitempty"`
	Ingredient  []*Edge   `json:"ingredient,omitempty"`
	Step        []*Edge   `json:"step,omitempty"`
	Conclusion  string    `json:"conclusion,omitempty"`
	Tag         []*Edge   `json:"tag,omitempty"`
	CreatedAt   time.Time `json:"createdAt,omitempty"`
	UpdatedAt   time.Time `json:"updatedAt,omitempty"`
}

type Edge struct {
	UID string `json:"uid,omitempty"`
}
