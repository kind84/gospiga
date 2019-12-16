package domain

type Recipe struct {
	id    int
	title string
}

func NewRecipe(id int, title string) (*Recipe, error) {
	r := &Recipe{
		id:    id,
		title: title,
	}
	return r, nil
}

func (r *Recipe) Hello() string {
	return r.title
}
