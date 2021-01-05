package gql

//go:generate go run github.com/99designs/gqlgen --verbose

import (
	"context"
	"errors"
	"strconv"

	"gospiga/pkg/types"
	"gospiga/server/domain"
)

func NewResolver(app App) *Resolver {
	return &Resolver{app: app}
}

type Resolver struct {
	app App
}

func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) Recipes(ctx context.Context, first *int, after *string,
	tags []*string, ingredients []*string, query *string) ([]*Recipe, error) {

	searchArgs := types.SearchRecipesArgs{
		First: first,
		After: after,
		Query: query,
	}

	if len(tags) > 0 {
		ts := make([]string, len(tags))
		for i, t := range tags {
			ts[i] = *t
		}
		searchArgs.Tags = ts
	}
	if len(ingredients) > 0 {
		is := make([]string, len(ingredients))
		for i, s := range ingredients {
			is[i] = *s
		}
		searchArgs.Ingredients = is
	}

	drr, err := r.app.SearchRecipes(ctx, &searchArgs)
	if err != nil {
		return nil, err
	}
	if len(drr) == 0 {
		return []*Recipe{}, nil
	}

	rr := make([]*Recipe, len(drr))
	for i, dr := range drr {
		rcp, err := mapRecipe(dr)
		if err != nil {
			return nil, err
		}
		rr[i] = rcp
	}

	return rr, nil
}

func mapRecipe(dr *domain.Recipe) (*Recipe, error) {
	var (
		mainImage  string
		cookTime   *int
		extraNotes *string
		conclusion *string
	)

	if dr.MainImage != nil {
		mainImage = dr.MainImage.URL
	}
	if dr.CookTime > 0 {
		cookTime = &[]int{dr.CookTime}[0]
	}
	if dr.ExtraNotes != "" {
		extraNotes = &[]string{dr.ExtraNotes}[0]
	}
	if dr.Conclusion != "" {
		conclusion = &[]string{dr.Conclusion}[0]
	}

	igrs := make([]*Ingredient, len(dr.Ingredients))
	for i, ding := range dr.Ingredients {
		var (
			uom *string
			qty *string
		)
		if ding.UnitOfMeasure != "" {
			uom = &[]string{ding.UnitOfMeasure}[0]
		}

		switch q := ding.Quantity.(type) {
		case string:
			qty = &q
		case int:
			qstr := strconv.Itoa(q)
			qty = &qstr
		default:
			return nil, errors.New("ingredient quantity type not valid")
		}

		ing := &Ingredient{
			Name:          ding.Name,
			UnitOfMeasure: uom,
			Quantity:      qty,
		}

		igrs[i] = ing
	}

	tags := make([]*Tag, len(dr.Tags))
	for t, dtag := range dr.Tags {
		tags[t] = &Tag{
			TagName: dtag.TagName,
		}
	}

	rcp := &Recipe{
		Xid:         dr.ExternalID,
		Title:       dr.Title,
		Subtitle:    dr.Subtitle,
		MainImage:   mainImage,
		Likes:       dr.Likes,
		Difficulty:  string(dr.Difficulty),
		Cost:        string(dr.Cost),
		PrepTime:    dr.PrepTime,
		CookTime:    cookTime,
		Servings:    dr.Servings,
		ExtraNotes:  extraNotes,
		Description: dr.Description,
		Ingredients: igrs,
		Conclusion:  conclusion,
		// TODO: map steps
		Tags: tags,
	}

	return rcp, nil
}
