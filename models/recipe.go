package models

import "time"

type Recipe struct {
	UID        string `json:"uid,omitempty"`
	Title      string `json:"title,omitempty"`
	Ingredient []struct {
		UID string `json:"uid,omitempty"`
	} `json:"ingredient,omitempty"`
	Step []struct {
		UID string `json:"uid,omitempty"`
	} `json:"step,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
}
