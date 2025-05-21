package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"gospiga/finder/domain"
	"gospiga/finder/fulltext"
	"gospiga/pkg/log"
	"gospiga/pkg/streamer"
	"gospiga/pkg/types"
)

func (a *app) SearchRecipes(query string) ([]*fulltext.Recipe, error) {
	return a.ft.SearchRecipes(query)
}

func (a *app) SearchByTag(tags []string) ([]*fulltext.Recipe, error) {
	return a.ft.SearchByTag(tags)
}

func (a *app) AllRecipeTags(ctx context.Context) ([]string, error) {
	return a.db.Tags(ctx, "recipes", "tags")
}

func (a *app) readNewRecipes(ctx context.Context) {
	msgChan := make(chan streamer.Message)
	var wg sync.WaitGroup

	streams := []string{
		savedRecipeStream,
		deletedRecipeStream,
	}
	args := &streamer.StreamArgs{
		Streams:  streams,
		Group:    group,
		Consumer: "finder-usecase",
		Messages: msgChan,
	}
	err := a.streamer.ReadGroup(ctx, &wg, args)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case msg := <-msgChan:
			switch msg.Stream {
			case savedRecipeStream:
				// ping-pong to parse recipe from message
				var recipe types.Recipe
				jr, err := json.Marshal(msg.Payload)
				if err != nil {
					log.Errorf("cannot read recipe ID from message ID %q", msg.ID)
					a.discardMessage(ctx, &msg, &wg)
					continue
				}
				err = json.Unmarshal(jr, &recipe)
				if err != nil {
					log.Errorf("cannot parse recipe ID from message ID %q", msg.ID)
					a.discardMessage(ctx, &msg, &wg)
					continue
				}
				log.Debugf("Got message for a saved recipe ID %q", recipe.ExternalID)

				a.indexRecipe(ctx, recipe, msg.Stream, msg.ID, &wg)

			case deletedRecipeStream:
				recipeID, ok := msg.Payload.(string)
				if !ok {
					log.Errorf("cannot read recipe ID from message ID %q", msg.ID)
					a.discardMessage(ctx, &msg, &wg)
					continue
				}
				log.Debugf("Got message for deleted recipe ID %q", recipeID)

				a.deleteRecipe(ctx, recipeID, msg.ID, &wg)
			}

		case <-ctx.Done():
			// time to exit
			return
		}
	}
}

func (a *app) indexRecipe(ctx context.Context, recipe types.Recipe, stream, messageID string, wg *sync.WaitGroup) {
	// unleash streamer
	defer wg.Done()

	// check if ID is already indexed
	if exists, _ := a.db.IDExists(fmt.Sprintf("recipe:%s", recipe.ID)); exists {
		log.Debugf("recipe ID %q already indexed", recipe.ID)

		err := a.streamer.Ack(ctx, stream, group, messageID)
		if err != nil {
			log.Errorf("error ack'ing msg ID %q", messageID)
			return
		}
	}

	r := domain.FromType(&recipe)

	// index recipe
	err := a.ft.IndexRecipe(r)
	if err != nil {
		log.Error(err)
		// TODO: ack??
		return
	}

	// ack (& add recipeIndexed?)
	err = a.streamer.Ack(ctx, stream, group, messageID)
	if err != nil {
		log.Errorf("error ack'ing msg ID %q", messageID)
	}
}

func (a *app) deleteRecipe(ctx context.Context, recipeID, messageID string, wg *sync.WaitGroup) {
	// unleash streamer
	defer wg.Done()

	err := a.ft.DeleteRecipe(recipeID)
	if err != nil {
		log.Errorf("error deleting recipe from index: %s", err)

		err := a.streamer.Ack(ctx, deletedRecipeStream, group, messageID)
		if err != nil {
			log.Errorf("error ack'ing msg ID %q", messageID)
		}
		return
	}

	err = a.streamer.Ack(ctx, deletedRecipeStream, group, messageID)
	if err != nil {
		log.Errorf("error ack'ing msg ID %q", messageID)
	}
}

func (a *app) discardMessage(ctx context.Context, m *streamer.Message, wg *sync.WaitGroup) {
	defer wg.Done()
	err := a.streamer.Ack(ctx, m.Stream, group, m.ID)
	if err != nil {
		log.Warnf("error acknowledging message: %s", err)
	}
}
