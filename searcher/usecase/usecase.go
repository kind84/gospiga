package usecase

import (
	"context"
	"fmt"

	gostreamer "github.com/kind84/gospiga/pkg/streamer"
	"github.com/kind84/gospiga/searcher/domain"
)

type App struct {
	ft       FT
	service  Service
	streamer Streamer
}

func NewApp(ctx context.Context, ft FT, service Service, streamer Streamer) *App {
	app := &App{
		ft:       ft,
		service:  service,
		streamer: streamer,
	}

	// start streamer to listen for new recipes.
	go app.readNewRecipes(ctx)

	return app
}

func (a *App) readNewRecipes(ctx context.Context) {
	msgChan := make(chan gostreamer.Message)
	exitChan := make(chan struct{})

	args := &gostreamer.StreamArgs{
		Stream:   "new-recipes",
		Group:    "usecase",
		Consumer: "usecase",
	}
	a.streamer.ReadGroup(ctx, args, msgChan, exitChan)

	for {
		select {
		case msg := <-msgChan:
			recipeID, ok := msg.Payload.(string)
			if !ok {
				fmt.Println("cannot read recipe ID from message.")
			}
			fmt.Printf("Got message for a new recipe ID %s\n", recipeID)

			// check if ID is already indexed

			// parse recipe from message
			var r *domain.Recipe

			// index recipe
			err := a.service.IndexRecipe(ctx, r)
			if err != nil {
				fmt.Println(err)
				continue
			}

			// ack & add recipeIndexed

		case <-ctx.Done():
			// time to exit
			close(exitChan)
		}
	}
}
