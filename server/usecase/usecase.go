package usecase

import (
	"context"

	log "github.com/sirupsen/logrus"

	gostreamer "github.com/kind84/gospiga/pkg/streamer"
)

type App struct {
	service  Service
	db       DB
	streamer Streamer
	provider Provider
}

func NewApp(ctx context.Context, service Service, db DB, streamer Streamer, provider Provider) *App {
	app := &App{
		service:  service,
		db:       db,
		streamer: streamer,
		provider: provider,
	}

	// start streamer to listen for new recipes.
	go app.readNewRecipes(ctx)

	return app
}

func (a *App) readNewRecipes(ctx context.Context) {
	msgChan := make(chan gostreamer.Message)
	exitChan := make(chan struct{})
	stream := "new-recipes"
	group := "server-usecase"

	args := &gostreamer.StreamArgs{
		Stream:   stream,
		Group:    group,
		Consumer: "usecase",
	}
	a.streamer.ReadGroup(ctx, args, msgChan, exitChan)

	for {
		select {
		case msg := <-msgChan:
			recipeID, ok := msg.Payload.(string)
			if !ok {
				log.Errorf("cannot read recipe ID from message ID [%s].", msg.ID)
				continue
			}
			log.Debugf("Got message for a new recipe ID [%s]", recipeID)

			// check if ID is already stored
			if saved, _ := a.service.IDSaved(ctx, recipeID); saved {
				log.Debugf("recipe ID [%s] already saved", recipeID)
				// ack HERE
				err := a.streamer.Ack(stream, group, msg.ID)
				if err != nil {
					log.Errorf("error ack'ing msg ID [%s]", msg.ID)
				}
				continue
			}

			// call datocms to get the full recipe
			r, err := a.provider.GetRecipe(ctx, recipeID)
			if err != nil {
				log.Error(err)
				continue
			}

			// save recipe
			err = a.service.SaveRecipe(ctx, r)
			if err != nil {
				log.Error(err)
				continue
			}

			rMsg := &gostreamer.Message{
				Payload: r,
			}

			err = a.streamer.AckAndAdd(args, "saved-recipes", msg.ID, rMsg)
			if err != nil {
				log.Errorf("error ack'ing msg ID [%s]", msg.ID)
			}

		case <-ctx.Done():
			// time to exit
			close(exitChan)
		}
	}
}
