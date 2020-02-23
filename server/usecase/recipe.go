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
func (a *app) NewRecipe(ctx context.Context, recipeID string) error {
	msg := &streamer.Message{Payload: recipeID}
	return a.streamer.Add(newRecipeStream, msg)
}

// UpdatedRecipe inform of an updated recipe ID sending it over the stream.
func (a *app) UpdatedRecipe(ctx context.Context, recipeID string) error {
	msg := &streamer.Message{Payload: recipeID}
	return a.streamer.Add(updatedRecipeStream, msg)
}

// DeletedRecipe inform of an deleted recipe ID sending it over the stream.
func (a *app) DeletedRecipe(ctx context.Context, recipeID string) error {
	msg := &streamer.Message{Payload: recipeID}
	return a.streamer.Add(deletedRecipeStream, msg)
}

// SearchRecipes matching the query string.
func (a *app) SearchRecipes(ctx context.Context, query string) ([]*domain.Recipe, error) {
	ids, err := a.stub.SearchRecipes(ctx, query)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return []*domain.Recipe{}, nil
	}

	return a.service.GetRecipesByIDs(ctx, ids)
}

func (a *app) readRecipes(ctx context.Context) error {
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
					recipeID, ok := msg.Payload.(string)
					if !ok {
						log.Errorf("cannot read recipe ID from message ID [%s].", msg.ID)
						// TODO: ack??
						wg.Done()
						continue
					}
					log.Debugf("Got message for a new recipe ID [%s]", recipeID)

					go a.upsertRecipe(ctx, recipeID, msg.Stream, msg.ID, &wg)

				case updatedRecipeStream:
					recipeID, ok := msg.Payload.(string)
					if !ok {
						log.Errorf("cannot read recipe ID from message ID [%s].", msg.ID)
						// TODO: ack??
						wg.Done()
						continue
					}
					log.Debugf("Got message for updated recipe ID [%s]", recipeID)

					go a.upsertRecipe(ctx, recipeID, msg.Stream, msg.ID, &wg)

				case deletedRecipeStream:
					recipeID, ok := msg.Payload.(string)
					if !ok {
						log.Errorf("cannot read recipe ID from message ID [%s].", msg.ID)
						// TODO: ack??
						wg.Done()
						continue
					}
					log.Debugf("Got message for deleted recipe ID [%s]", recipeID)

					go a.deleteRecipe(ctx, recipeID, msg.ID, &wg)

				}

			case <-ctx.Done():
				// time to exit
				close(exitChan)
			}
		}
	}()
	return nil
}

func (a *app) upsertRecipe(ctx context.Context, recipeID, fromStream, messageID string, wg *sync.WaitGroup) {
	// call provider to get the full recipe
	r, err := a.provider.GetRecipe(ctx, recipeID)
	if err != nil {
		log.Error(err)
		// TODO: ack?? new stream??
		wg.Done()
		return
	}

	// save recipe
	err = a.service.SaveRecipe(ctx, r)
	if err != nil {
		log.Error(err)
		// TODO: ack ??
		wg.Done()
		return
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
}

func (a *app) deleteRecipe(ctx context.Context, recipeID, messageID string, wg *sync.WaitGroup) {
	// delete recipe
	err := a.service.DeleteRecipe(ctx, recipeID)
	if err != nil {
		log.Error(err)
		// TODO: ack ??
		wg.Done()
		return
	}

	// ack message
	// TODO: relay on deleted-stream??
	err = a.streamer.Ack(deletedRecipeStream, group, messageID)
	if err != nil {
		log.Errorf("error on Ack for msg ID [%s]", messageID)
	}

	// unleash the streamer
	wg.Done()
}
