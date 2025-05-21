package usecase

import "context"

const (
	savedRecipeStream   = "saved-recipes"
	deletedRecipeStream = "deleted-recipes"
	group               = "finder-usecase"
)

type app struct {
	db       DB
	ft       FT
	streamer Streamer
}

func NewApp(ctx context.Context, db DB, ft FT, streamer Streamer) *app {
	a := &app{
		db:       db,
		ft:       ft,
		streamer: streamer,
	}

	// start streamer to listen for new recipes.
	go a.readNewRecipes(ctx)

	return a
}
