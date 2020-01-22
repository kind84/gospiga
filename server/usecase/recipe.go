package usecase

import (
	"context"
	"sync"

	"github.com/kind84/gospiga/pkg/streamer"
	"github.com/kind84/gospiga/server/domain"
	log "github.com/sirupsen/logrus"

	gostreamer "github.com/kind84/gospiga/pkg/streamer"
)

const (
	newRecipeStream     = "new-recipes"
	updatedRecipeStream = "updated-recipes"
	deletedRecipeStream = "deleted-recipes"
	group               = "server-usecase"
)

// NewRecipe inform of a new recipe ID sending it over the stream.
func (a *App) NewRecipe(ctx context.Context, recipeID string) error {
	msg := &streamer.Message{Payload: recipeID}
	return a.streamer.Add(newRecipeStream, msg)
}

// UpdatedRecipe inform of an updated recipe ID sending it over the stream.
func (a *App) UpdatedRecipe(ctx context.Context, recipeID string) error {
	msg := &streamer.Message{Payload: recipeID}
	return a.streamer.Add(updatedRecipeStream, msg)
}

// DeletedRecipe inform of an deleted recipe ID sending it over the stream.
func (a *App) DeletedRecipe(ctx context.Context, recipeID string) error {
	msg := &streamer.Message{Payload: recipeID}
	return a.streamer.Add(deletedRecipeStream, msg)
}

// SearchRecipes matching the query string.
func (a *App) SearchRecipes(ctx context.Context, query string) ([]*domain.Recipe, error) {
	ids, err := a.stub.SearchRecipes(ctx, query)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return []*domain.Recipe{}, nil
	}

	return a.service.GetRecipesByIDs(ctx, ids)
}

func (a *App) readRecipes(ctx context.Context) error {
	msgChan := make(chan gostreamer.Message)
	exitChan := make(chan struct{})
	var wg sync.WaitGroup

	streams := []string{
		newRecipeStream,
		updatedRecipeStream,
		deletedRecipeStream,
	}
	args := &gostreamer.StreamArgs{
		Streams:  streams,
		Group:    group,
		Consumer: "usecase",
	}

	err := a.streamer.ReadGroup(ctx, args, msgChan, exitChan, &wg)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case msg := <-msgChan:
				switch msg.Stream {
				case newRecipeStream:
					recipeID, ok := msg.Payload.(string)
					if !ok {
						log.Errorf("cannot read recipe ID from message ID [%s].", msg.ID)
						continue
					}
					log.Debugf("Got message for a new recipe ID [%s]", recipeID)

					a.ProcessNewRecipe(ctx, recipeID, msg.Stream, msg.ID, &wg)

				case updatedRecipeStream:
					recipeID, ok := msg.Payload.(string)
					if !ok {
						log.Errorf("cannot read recipe ID from message ID [%s].", msg.ID)
						continue
					}
					log.Debugf("Got message for a new recipe ID [%s]", recipeID)

					a.ProcessRecipeUpdate(ctx, recipeID, msg.Stream, msg.ID, &wg)
				}

			case <-ctx.Done():
				// time to exit
				close(exitChan)
			}
		}
	}()
	return nil
}

func (a *App) ProcessNewRecipe(ctx context.Context, recipeID, fromStream, messageID string, wg *sync.WaitGroup) {
	go func() {
		// check if recipe is already stored
		if r, err := a.service.GetRecipeByID(ctx, recipeID); err != nil && r != nil {
			log.Debugf("recipe ID [%s] already saved", recipeID)

			rMsg := &gostreamer.Message{
				Payload: r,
			}

			err = a.streamer.AckAndAdd(fromStream, "saved-recipes", group, messageID, rMsg)
			if err != nil {
				log.Errorf("error on AckAndAdd for msg ID [%s]", messageID)
			}
		}

		// call provider to get the full recipe
		r, err := a.provider.GetRecipe(ctx, recipeID)
		if err != nil {
			log.Error(err)
		}

		// save recipe
		err = a.service.SaveRecipe(ctx, r)
		if err != nil {
			log.Error(err)
		}

		// ack message and relay
		rMsg := &gostreamer.Message{
			Payload: r,
		}
		err = a.streamer.AckAndAdd(fromStream, "saved-recipes", group, messageID, rMsg)
		if err != nil {
			log.Errorf("error on AckAndAdd for msg ID [%s]", messageID)
		}

		// unleash the streamer
		wg.Done()
	}()
}

func (a *App) ProcessRecipeUpdate(ctx context.Context, recipeID, fromStream, messageID string, wg *sync.WaitGroup) {
	go func() {
		// call provider to get the full recipe
		r, err := a.provider.GetRecipe(ctx, recipeID)
		if err != nil {
			log.Error(err)
		}

		// update recipe
		err = a.service.UpdateRecipe(ctx, r)
		if err != nil {
			log.Error(err)
		}

		// ack message
		err = a.streamer.Ack(fromStream, group, messageID)
		if err != nil {
			log.Errorf("error on Ack for msg ID [%s]", messageID)
		}

		// unleash the streamer
		wg.Done()
	}()
}

func (a *App) ProcessDeleteRecipe(ctx context.Context, recipeID, fromStream, messageID string, wg *sync.WaitGroup) {
	go func() {
		// delete recipe
		err := a.service.DeleteRecipe(ctx, recipeID)
		if err != nil {
			log.Error(err)
		}

		// ack message
		err = a.streamer.Ack(fromStream, group, messageID)
		if err != nil {
			log.Errorf("error on Ack for msg ID [%s]", messageID)
		}

		// unleash the streamer
		wg.Done()
	}()
}
