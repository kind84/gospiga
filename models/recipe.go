package models

import "time"

type Recipe struct {
	UID         string `json:"uid,omitempty"`
	Title       string `json:"title,omitempty"`
	Subtitle    string `json:"subtitle,omitempty"`
	Description string `json:"description,omitempty"`
	Ingredient  []struct {
		UID string `json:"uid,omitempty"`
	} `json:"ingredient,omitempty"`
	Step []struct {
		UID string `json:"uid,omitempty"`
	} `json:"step,omitempty"`
	Conclusion string    `json:"conclusion,omitempty"`
	CreatedAt  time.Time `json:"createdAt,omitempty"`
	UpdatedAt  time.Time `json:"updatedAt,omitempty"`
}
