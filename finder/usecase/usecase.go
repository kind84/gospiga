package usecase

const (
	savedRecipeStream   = "saved-recipes"
	deletedRecipeStream = "deleted-recipes"
	group               = "finder-usecase"
)

type app struct {
	db       DB
	ft       FT
	streamer Streamer
	shutdown chan struct{}
}

// CloseGracefully sends the shutdown signal to start closing all app processes
func (a *app) CloseGracefully() {
	close(a.shutdown)
}

func NewApp(db DB, ft FT, streamer Streamer) *app {
	a := &app{
		db:       db,
		ft:       ft,
		streamer: streamer,
		shutdown: make(chan struct{}),
	}

	// start streamer to listen for new recipes.
	go a.readNewRecipes()

	return a
}
