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
	Conclusion string `json:"conclusion,omitempty"`
	Tag        []struct {
		UID string `json:"uid,omitempty"`
	} `json:"tag,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`
}

//
// type Step struct {
// 	UID     string `json:"uid,omitempty"`
// 	Excerpt string `json:"excerpt,omitempty"`
// 	Text    string `json:"text,omitempty"`
// 	Index   int    `json:"index,omitempty"`
// }
//
// type Tag struct {
// 	UID  string `json:"uid,omitempty"`
// 	Name string `json:"name,omitempty"`
// }
