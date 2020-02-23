package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/kind84/gospiga/finder/domain"
	"github.com/kind84/gospiga/pkg/streamer"
	"github.com/kind84/gospiga/pkg/types"
)

const (
	newRecipeStream     = "new-recipes"
	updatedRecipeStream = "updated-recipes"
	deletedRecipeStream = "deleted-recipes"
	group               = "finder-usecase"
)

type app struct {
	db       DB
	ft       FT
	streamer Streamer
}

func NewApp(ctx context.Context, db DB, ft FT, streamer Streamer) (*app, error) {
	a := &app{
		db:       db,
		ft:       ft,
		streamer: streamer,
	}

	// start streamer to listen for new recipes.
	err := a.readNewRecipes(ctx)
	if err != nil {
		return nil, err
	}

	return a, nil
}

func (a *app) readNewRecipes(ctx context.Context) error {
	msgChan := make(chan streamer.Message)
	exitChan := make(chan struct{})
	var wg sync.WaitGroup

	streams := []string{
		newRecipeStream,
		updatedRecipeStream,
		deletedRecipeStream,
	}
	args := &streamer.StreamArgs{
		Streams:  streams,
		Group:    group,
		Consumer: "finder-usecase",
		Messages: msgChan,
		Exit:     exitChan,
		WG:       &wg,
	}
	err := a.streamer.ReadGroup(ctx, args)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case msg := <-msgChan:
				switch msg.Stream {
				case newRecipeStream:
					// ping-pong to parse recipe from message
					var recipeRaw types.Recipe
					jr, err := json.Marshal(msg.Payload)
					if err != nil {
						log.Errorf("cannot read recipe ID from message ID [%s].", msg.ID)
						a.discardMessage(&msg, &wg)
						continue
					}
					err = json.Unmarshal(jr, &recipeRaw)
					if err != nil {
						log.Errorf("cannot parse recipe ID from message ID [%s].", msg.ID)
						a.discardMessage(&msg, &wg)
						continue
					}
					log.Debugf("Got message for a new recipe ID [%s]", recipeRaw.ID)

					go a.indexRecipe(recipeRaw, msg.Stream, msg.ID, &wg)

				case updatedRecipeStream:
					// ping-pong to parse recipe from message
					var recipeRaw types.Recipe
					jr, err := json.Marshal(msg.Payload)
					if err != nil {
						log.Errorf("cannot read recipe ID from message ID [%s].", msg.ID)
						a.discardMessage(&msg, &wg)
						continue
					}
					err = json.Unmarshal(jr, &recipeRaw)
					if err != nil {
						log.Errorf("cannot parse recipe ID from message ID [%s].", msg.ID)
						a.discardMessage(&msg, &wg)
						continue
					}
					log.Debugf("Got message for updated recipe ID [%s]", recipeRaw.ID)

					go a.indexRecipe(recipeRaw, msg.Stream, msg.ID, &wg)

				case deletedRecipeStream:
					recipeID, ok := msg.Payload.(string)
					if !ok {
						log.Errorf("cannot read recipe ID from message ID [%s].", msg.ID)
						a.discardMessage(&msg, &wg)
						continue
					}
					log.Debugf("Got message for deleted recipe ID [%s]", recipeID)

					go a.deleteRecipe(recipeID, msg.ID, &wg)
				}

			case <-ctx.Done():
				// time to exit
				close(exitChan)
			}
		}
	}()
	return nil
}

func (a *app) indexRecipe(recipeRaw types.Recipe, stream, messageID string, wg *sync.WaitGroup) {
	// check if ID is already indexed
	if exists, _ := a.db.IDExists(fmt.Sprintf("recipe-%s", recipeRaw.ID)); exists {
		log.Debugf("recipe ID [%s] already indexed", recipeRaw.ID)

		err := a.streamer.Ack(stream, group, messageID)
		if err != nil {
			log.Errorf("error ack'ing msg ID [%s]", messageID)
			wg.Done()
			return
		}
	}

	r := domain.MapFromType(&recipeRaw)

	// index recipe
	err := a.ft.IndexRecipe(r)
	if err != nil {
		log.Error(err)
		// TODO: ack??
		wg.Done()
		return
	}

	// ack (& add recipeIndexed?)
	err = a.streamer.Ack(stream, group, messageID)
	if err != nil {
		log.Errorf("error ack'ing msg ID [%s]", messageID)
	}

	// unleash streamer
	wg.Done()
}

func (a *app) deleteRecipe(recipeID, messageID string, wg *sync.WaitGroup) {
	err := a.ft.DeleteRecipe(recipeID)
	if err != nil {
		log.Errorf("error deleting recipe from index: %s", err)

		err := a.streamer.Ack(deletedRecipeStream, group, messageID)
		if err != nil {
			log.Errorf("error ack'ing msg ID [%s]", messageID)
		}
		wg.Done()
		return
	}

	err = a.streamer.Ack(deletedRecipeStream, group, messageID)
	if err != nil {
		log.Errorf("error ack'ing msg ID [%s]", messageID)
	}

	// unleash streamer
	wg.Done()
}

func (a *app) discardMessage(m *streamer.Message, wg *sync.WaitGroup) {
	err := a.streamer.Ack(m.Stream, group, m.ID)
	if err != nil {
		log.Warnf("error acknowledging message: %s", err)
	}
	wg.Done()
}
