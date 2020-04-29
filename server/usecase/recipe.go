package usecase

import (
	"context"
	"errors"
	"sync"

	"github.com/kind84/gospiga/server/domain"
	log "github.com/sirupsen/logrus"

	errs "github.com/kind84/gospiga/pkg/errors"
	"github.com/kind84/gospiga/pkg/streamer"
)

const (
	newRecipeStream     = "new-recipes"
	updatedRecipeStream = "updated-recipes"
	deletedRecipeStream = "deleted-recipes"
	group               = "server-usecase"
)

// NewRecipe informs of a new recipe ID sending it over the stream.
func (a *app) NewRecipe(ctx context.Context, recipeID string) error {
	return a.streamer.Add(newRecipeStream, &streamer.Message{Payload: recipeID})
}

// UpdatedRecipe informs of an updated recipe ID sending it over the stream.
func (a *app) UpdatedRecipe(ctx context.Context, recipeID string) error {
	return a.streamer.Add(updatedRecipeStream, &streamer.Message{Payload: recipeID})
}

// DeletedRecipe informs of an deleted recipe ID sending it over the stream.
func (a *app) DeletedRecipe(ctx context.Context, recipeID string) error {
	return a.streamer.Add(deletedRecipeStream, &streamer.Message{Payload: recipeID})
}

// RecipeTags returns the set of used tags.
func (a *app) RecipeTags(ctx context.Context) ([]string, error) {
	tags, err := a.stub.AllRecipeTags(ctx)
	if err != nil {
		return nil, err
	}

	return tags, nil
}

// LoadRecipes in the platform by injecting all the recipe IDs retrieved from
// the provider.
func (a *app) LoadRecipes(ctx context.Context) error {
	rids, err := a.provider.GetAllRecipeIDs(ctx)
	if err != nil {
		return err
	}

	for _, id := range rids {
		err := a.streamer.Add(newRecipeStream, &streamer.Message{Payload: id})
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *app) readRecipes() {
	ctx, exit := context.WithCancel(context.Background())
	msgChan := make(chan streamer.Message)
	var wg sync.WaitGroup

	streams := []string{
		newRecipeStream,
		updatedRecipeStream,
		deletedRecipeStream,
	}
	args := &streamer.StreamArgs{
		Streams:  streams,
		Group:    group,
		Consumer: "usecase",
		Messages: msgChan,
		Exit:     a.shutdown,
		WG:       &wg,
	}

	err := a.streamer.ReadGroup(args)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case msg := <-msgChan:
			switch msg.Stream {
			case newRecipeStream:
				recipeID, ok := msg.Payload.(string)
				if !ok {
					log.Errorf("cannot read recipe ID from message ID %q", msg.ID)
					a.discardMessage(&msg, &wg)
					continue
				}
				log.Debugf("Got message for a new recipe ID %q", recipeID)

				a.saveRecipe(ctx, recipeID, msg.Stream, msg.ID, &wg)

			case updatedRecipeStream:
				recipeID, ok := msg.Payload.(string)
				if !ok {
					log.Errorf("cannot read recipe ID from message ID %q", msg.ID)
					a.discardMessage(&msg, &wg)
					continue
				}
				log.Debugf("Got message for updated recipe ID %q", recipeID)

				a.upsertRecipe(ctx, recipeID, msg.Stream, msg.ID, &wg)

			case deletedRecipeStream:
				recipeID, ok := msg.Payload.(string)
				if !ok {
					log.Errorf("cannot read recipe ID from message ID %q", msg.ID)
					a.discardMessage(&msg, &wg)
					continue
				}
				log.Debugf("Got message for deleted recipe ID %q", recipeID)

				a.deleteRecipe(ctx, recipeID, msg.ID, &wg)

			}

		case <-a.shutdown:
			// time to exit
			exit()
			return
		}
	}
}

func (a *app) saveRecipe(ctx context.Context, recipeID, fromStream, messageID string, wg *sync.WaitGroup) {
	// call provider to get the full recipe
	rt, err := a.provider.GetRecipe(ctx, recipeID)
	if err != nil {
		log.Error(err)
		// TODO: ack?? new stream??
		wg.Done()
		return
	}
	r := domain.FromType(rt)

	// save recipe
	err = a.service.SaveRecipe(ctx, r)
	var errdup errs.ErrDuplicateID
	if errors.As(err, &errdup) {
		log.Infof("recipe ID %q already saved", r.ExternalID)
		err = a.streamer.Ack(fromStream, group, messageID)
		if err != nil {
			log.Errorf("error on Ack for msg ID %q", messageID)
		}
		wg.Done()
		return
	}
	if err != nil {
		log.Error(err)
		// TODO: ack ??
		wg.Done()
		return
	}

	// ack message and relay
	rMsg := &streamer.Message{
		Payload: r.ToType(),
	}
	err = a.streamer.AckAndAdd(fromStream, "saved-recipes", group, messageID, rMsg)
	if err != nil {
		log.Errorf("error on AckAndAdd for msg ID %q", messageID)
	}

	// unleash the streamer
	wg.Done()
}

func (a *app) upsertRecipe(ctx context.Context, recipeID, fromStream, messageID string, wg *sync.WaitGroup) {
	// call provider to get the full recipe
	rt, err := a.provider.GetRecipe(ctx, recipeID)
	r := domain.FromType(rt)
	if err != nil {
		log.Error(err)
		// TODO: ack?? new stream??
		wg.Done()
		return
	}

	// save recipe
	err = a.service.UpdateRecipe(ctx, r)
	if err != nil {
		log.Error(err)
		// TODO: ack ??
		wg.Done()
		return
	}

	// ack message and relay
	rMsg := &streamer.Message{
		Payload: r.ToType(),
	}
	err = a.streamer.AckAndAdd(fromStream, "saved-recipes", group, messageID, rMsg)
	if err != nil {
		log.Errorf("error on AckAndAdd for msg ID %q", messageID)
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

	// TODO: relay on deleted-stream??
	err = a.streamer.Ack(deletedRecipeStream, group, messageID)
	if err != nil {
		log.Errorf("error on Ack for msg ID %q", messageID)
	}

	// unleash the streamer
	wg.Done()
}

func (a *app) discardMessage(m *streamer.Message, wg *sync.WaitGroup) {
	err := a.streamer.Ack(m.Stream, group, m.ID)
	if err != nil {
		log.Warnf("error acknowledging message: %s", err)
	}
	wg.Done()
}
