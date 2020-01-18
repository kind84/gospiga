package usecase

import (
	"context"
	"sync"

	log "github.com/sirupsen/logrus"

	gostreamer "github.com/kind84/gospiga/pkg/streamer"
)

type App struct {
	service  Service
	db       DB
	streamer Streamer
	provider Provider
	stub     Stub
}

func NewApp(ctx context.Context, service Service, db DB, streamer Streamer, provider Provider, stub Stub) *App {
	app := &App{
		service:  service,
		db:       db,
		streamer: streamer,
		provider: provider,
		stub:     stub,
	}

	// start streamer to listen for new recipes.
	go app.readNewRecipes(ctx)

	return app
}

func (a *App) readNewRecipes(ctx context.Context) {
	msgChan := make(chan gostreamer.Message)
	exitChan := make(chan struct{})
	var wg sync.WaitGroup
	stream := "new-recipes"
	group := "server-usecase"

	args := &gostreamer.StreamArgs{
		Stream:   stream,
		Group:    group,
		Consumer: "usecase",
	}
	a.streamer.ReadGroup(ctx, args, msgChan, exitChan, &wg)

	for {
		select {
		case msg := <-msgChan:
			recipeID, ok := msg.Payload.(string)
			if !ok {
				log.Errorf("cannot read recipe ID from message ID [%s].", msg.ID)
				continue
			}
			log.Debugf("Got message for a new recipe ID [%s]", recipeID)

			// check if recipe is already stored
			if r, err := a.service.GetRecipeByID(ctx, recipeID); err != nil && r != nil {
				log.Debugf("recipe ID [%s] already saved", recipeID)

				rMsg := &gostreamer.Message{
					Payload: r,
				}

				err = a.streamer.AckAndAdd(args, "saved-recipes", msg.ID, rMsg)
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

			// ack message and relay
			rMsg := &gostreamer.Message{
				Payload: r,
			}
			err = a.streamer.AckAndAdd(args, "saved-recipes", msg.ID, rMsg)
			if err != nil {
				log.Errorf("error on AckAndAdd for msg ID [%s]", msg.ID)
				continue
			}

			// unleash streamer
			wg.Done()

		case <-ctx.Done():
			// time to exit
			close(exitChan)
		}
	}
}
