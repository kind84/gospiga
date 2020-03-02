package usecase

type app struct {
	service  Service
	db       DB
	streamer Streamer
	provider Provider
	stub     Stub
	shutdown chan struct{}
}

func NewApp(service Service, db DB, streamer Streamer, provider Provider, stub Stub) *app {
	a := &app{
		service:  service,
		db:       db,
		streamer: streamer,
		provider: provider,
		stub:     stub,
		shutdown: make(chan struct{}),
	}

	// start streamer to listen for new recipes.
	go a.readRecipes()

	return a
}

// CloseGracefully sends the shutdown signal to start closing all app processes
func (a *app) CloseGracefully() {
	close(a.shutdown)
}
