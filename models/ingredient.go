package models

type Ingredient struct {
	UID      string `json:"uid,omitempty"`
	Name     string `json:"name,omitempty"`
	Quantity int    `json:"quantity,omitempty"`
	Recipe   []struct {
		UID string `json:"uid,omitempty"`
	} `json:"recipe,omitempty"`
}
