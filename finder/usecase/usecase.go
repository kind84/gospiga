package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"

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
				log.Errorf("cannot parse recipe from message: %s\n", err)
			}
			log.Debugf("Got message for a new recipe ID %s\n", recipeRaw.ID)

			// check if ID is already indexed
			if exists, _ := a.db.IDExists(fmt.Sprintf("recipe-%s", recipeRaw.ID)); exists {
				log.Debugf("recipe ID [%s] already indexed", recipeRaw.ID)

				err := a.streamer.Ack(stream, group, msg.ID)
				if err != nil {
					log.Errorf("error ack'ing msg ID %s\n", msg.ID)
				}
				continue
			}

			var r *domain.Recipe
			r.MapFromType(&recipeRaw)

			// index recipe
			err = a.ft.IndexRecipe(r)
			if err != nil {
				log.Error(err)
				continue
			}

			// ack (& add recipeIndexed?)
			err = a.streamer.Ack(stream, group, msg.ID)
			if err != nil {
				log.Errorf("error ack'ing msg ID %s\n", msg.ID)
			}

		case <-ctx.Done():
			// time to exit
			close(exitChan)
		}
	}
}
