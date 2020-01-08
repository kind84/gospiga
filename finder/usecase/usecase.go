package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kind84/gospiga/finder/domain"
	gostreamer "github.com/kind84/gospiga/pkg/streamer"
	"github.com/kind84/gospiga/pkg/types"
)

type App struct {
	db       DB
	ft       FT
	streamer Streamer
}

func NewApp(ctx context.Context, db DB, ft FT, streamer Streamer) *App {
	app := &App{
		db:       db,
		ft:       ft,
		streamer: streamer,
	}

	// start streamer to listen for new recipes.
	go app.readNewRecipes(ctx)

	return app
}

func (a *App) readNewRecipes(ctx context.Context) {
	msgChan := make(chan gostreamer.Message)
	exitChan := make(chan struct{})
	stream := "saved-recipes"
	group := "finder-usecase"

	args := &gostreamer.StreamArgs{
		Stream:   stream,
		Group:    group,
		Consumer: "finder-usecase",
	}
	a.streamer.ReadGroup(ctx, args, msgChan, exitChan)

	for {
		select {
		case msg := <-msgChan:
			// parse recipe from message
			var recipeRaw types.Recipe
			err := json.Unmarshal(msg.Payload.([]byte), &recipeRaw)
			if err != nil {
				fmt.Printf("cannot parse recipe from message: %s\n", err)
			}
			fmt.Printf("Got message for a new recipe ID %s\n", recipeRaw.ID)

			// check if ID is already indexed
			if exists, _ := a.db.IDExists(fmt.Sprintf("recipe-%s", recipeRaw.ID)); exists {
				fmt.Printf("recipe ID [%s] already indexed", recipeRaw.ID)

				err := a.streamer.Ack(stream, group, msg.ID)
				if err != nil {
					fmt.Printf("error ack'ing msg ID %s\n", msg.ID)
				}
				continue
			}

			var r *domain.Recipe
			r.MapFromType(&recipeRaw)

			// index recipe
			err = a.ft.IndexRecipe(r)
			if err != nil {
				fmt.Println(err)
				continue
			}

			// ack (& add recipeIndexed?)
			err = a.streamer.Ack(stream, group, msg.ID)
			if err != nil {
				fmt.Printf("error ack'ing msg ID %s\n", msg.ID)
			}

		case <-ctx.Done():
			// time to exit
			close(exitChan)
		}
	}
}
