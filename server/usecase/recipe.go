package usecase

import (
	"context"
	"errors"
	"sync"

	errs "gospiga/pkg/errors"
	"gospiga/pkg/log"
	"gospiga/pkg/streamer"
	"gospiga/server/domain"
)

const (
	newRecipeStream     = "new-recipes"
	updatedRecipeStream = "updated-recipes"
	deletedRecipeStream = "deleted-recipes"
	group               = "server-usecase"
)

// NewRecipe informs of a new recipe ID sending it over the stream.
func (a *app) NewRecipe(ctx context.Context, recipeID string) error {
	return a.streamer.Add(ctx, newRecipeStream, &streamer.Message{Payload: recipeID})
}

// UpdatedRecipe informs of an updated recipe ID sending it over the stream.
func (a *app) UpdatedRecipe(ctx context.Context, recipeID string) error {
	return a.streamer.Add(ctx, updatedRecipeStream, &streamer.Message{Payload: recipeID})
}

// DeletedRecipe informs of an deleted recipe ID sending it over the stream.
func (a *app) DeletedRecipe(ctx context.Context, recipeID string) error {
	return a.streamer.Add(ctx, deletedRecipeStream, &streamer.Message{Payload: recipeID})
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
		err := a.streamer.Add(ctx, newRecipeStream, &streamer.Message{Payload: id})
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *app) readRecipes(ctx context.Context) {
	msgChan := make(chan streamer.Message)
	var wg sync.WaitGroup

	streams := []string{
		newRecipeStream,
		updatedRecipeStream,
		deletedRecipeStream,
	}
	args := streamer.StreamArgs{
		Streams:  streams,
		Group:    group,
		Consumer: "usecase",
		Messages: msgChan,
	}

	err := a.streamer.ReadGroup(ctx, &wg, &args)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case msg := <-msgChan:
			recipeID, ok := msg.Payload.(string)
			if !ok {
				log.Errorf("cannot read recipe ID from message ID %q", msg.ID)
				a.discardMessage(ctx, &msg, &wg)
				continue
			}

			switch msg.Stream {
			case newRecipeStream:
				log.Debugf("Got message for a new recipe ID %q", recipeID)

				a.saveRecipe(ctx, recipeID, msg.Stream, msg.ID, &wg)

			case updatedRecipeStream:
				log.Debugf("Got message for updated recipe ID %q", recipeID)

				a.updateRecipe(ctx, recipeID, msg.Stream, msg.ID, &wg)

			case deletedRecipeStream:
				log.Debugf("Got message for deleted recipe ID %q", recipeID)

				a.deleteRecipe(ctx, recipeID, msg.ID, &wg)
			}

		case <-ctx.Done():
			// time to exit
			return
		}
	}
}

func (a *app) saveRecipe(ctx context.Context, recipeID, fromStream, messageID string, wg *sync.WaitGroup) {
	// unleash the streamer
	defer wg.Done()

	// call provider to get the full recipe
	rt, err := a.provider.GetRecipe(ctx, recipeID)
	if err != nil {
		log.Error(err)
		// TODO: ack?? new stream??
		return
	}
	r := domain.FromType(rt)

	// save recipe
	err = a.service.SaveRecipe(ctx, r)
	var errdup errs.ErrDuplicateID
	if errors.As(err, &errdup) {
		log.Infof("recipe ID %q already saved", r.ExternalID)
		err = a.streamer.Ack(ctx, fromStream, group, messageID)
		if err != nil {
			log.Errorf("error on Ack for msg ID %q", messageID)
		}
		return
	}
	if err != nil {
		log.Error(err)
		// TODO: ack ??
		return
	}

	// ack message and relay
	rMsg := &streamer.Message{
		Payload: r.ToType(),
	}
	err = a.streamer.AckAndAdd(ctx, fromStream, "saved-recipes", group, messageID, rMsg)
	if err != nil {
		log.Errorf("error on AckAndAdd for msg ID %q", messageID)
	}
}

func (a *app) updateRecipe(ctx context.Context, recipeID, fromStream, messageID string, wg *sync.WaitGroup) {
	// unleash the streamer
	defer wg.Done()

	// call provider to get the full recipe
	rt, err := a.provider.GetRecipe(ctx, recipeID)
	r := domain.FromType(rt)
	if err != nil {
		log.Error(err)
		// TODO: ack?? new stream??
		return
	}

	// save recipe
	rID, err := a.service.UpdateRecipe(ctx, r)
	if err != nil {
		log.Error(err)
		// TODO: ack ??
		return
	}
	if rID != "" {
		r.ID = rID
	}

	// ack message and relay
	rMsg := &streamer.Message{
		Payload: r.ToType(),
	}
	err = a.streamer.AckAndAdd(ctx, fromStream, "saved-recipes", group, messageID, rMsg)
	if err != nil {
		log.Errorf("error on AckAndAdd for msg ID %q", messageID)
	}
}

func (a *app) deleteRecipe(ctx context.Context, recipeID, messageID string, wg *sync.WaitGroup) {
	// unleash the streamer
	defer wg.Done()

	// delete recipe
	err := a.service.DeleteRecipe(ctx, recipeID)
	if err != nil {
		log.Error(err)
		// TODO: ack ??
		return
	}

	// TODO: relay on deleted-stream??
	err = a.streamer.Ack(ctx, deletedRecipeStream, group, messageID)
	if err != nil {
		log.Errorf("error on Ack for msg ID %q", messageID)
	}
}

func (a *app) discardMessage(ctx context.Context, m *streamer.Message, wg *sync.WaitGroup) {
	defer wg.Done()
	err := a.streamer.Ack(ctx, m.Stream, group, m.ID)
	if err != nil {
		log.Warnf("error acknowledging message: %s", err)
	}
}
