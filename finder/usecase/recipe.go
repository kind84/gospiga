package usecase

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/kind84/gospiga/finder/domain"
	"github.com/kind84/gospiga/finder/fulltext"
	"github.com/kind84/gospiga/pkg/log"
	"github.com/kind84/gospiga/pkg/streamer"
	"github.com/kind84/gospiga/pkg/types"
)

func (a *app) SearchRecipes(query string) ([]*fulltext.Recipe, error) {
	return a.ft.SearchRecipes(query)
}

func (a *app) SearchByTag(tags []string) ([]*fulltext.Recipe, error) {
	return a.ft.SearchByTag(tags)
}

func (a *app) AllRecipeTags() ([]string, error) {
	return a.db.Tags("recipes", "tags")
}

func (a *app) readNewRecipes() {
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
			case savedRecipeStream:
				// ping-pong to parse recipe from message
				var recipe types.Recipe
				jr, err := json.Marshal(msg.Payload)
				if err != nil {
					log.Errorf("cannot read recipe ID from message ID %q", msg.ID)
					a.discardMessage(&msg, &wg)
					continue
				}
				err = json.Unmarshal(jr, &recipe)
				if err != nil {
					log.Errorf("cannot parse recipe ID from message ID %q", msg.ID)
					a.discardMessage(&msg, &wg)
					continue
				}
				log.Debugf("Got message for a saved recipe ID %q", recipe.ExternalID)

				a.indexRecipe(recipe, msg.Stream, msg.ID, &wg)

			case deletedRecipeStream:
				recipeID, ok := msg.Payload.(string)
				if !ok {
					log.Errorf("cannot read recipe ID from message ID %q", msg.ID)
					a.discardMessage(&msg, &wg)
					continue
				}
				log.Debugf("Got message for deleted recipe ID %q", recipeID)

				a.deleteRecipe(recipeID, msg.ID, &wg)
			}

		case <-a.shutdown:
			// time to exit
			return
		}
	}
}

func (a *app) indexRecipe(recipe types.Recipe, stream, messageID string, wg *sync.WaitGroup) {
	// unleash streamer
	defer wg.Done()

	// check if ID is already indexed
	if exists, _ := a.db.IDExists(fmt.Sprintf("recipe:%s", recipe.ID)); exists {
		log.Debugf("recipe ID %q already indexed", recipe.ID)

		err := a.streamer.Ack(stream, group, messageID)
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
	err = a.streamer.Ack(stream, group, messageID)
	if err != nil {
		log.Errorf("error ack'ing msg ID %q", messageID)
	}
}

func (a *app) deleteRecipe(recipeID, messageID string, wg *sync.WaitGroup) {
	// unleash streamer
	defer wg.Done()

	err := a.ft.DeleteRecipe(recipeID)
	if err != nil {
		log.Errorf("error deleting recipe from index: %s", err)

		err := a.streamer.Ack(deletedRecipeStream, group, messageID)
		if err != nil {
			log.Errorf("error ack'ing msg ID %q", messageID)
		}
		return
	}

	err = a.streamer.Ack(deletedRecipeStream, group, messageID)
	if err != nil {
		log.Errorf("error ack'ing msg ID %q", messageID)
	}
}

func (a *app) discardMessage(m *streamer.Message, wg *sync.WaitGroup) {
	defer wg.Done()
	err := a.streamer.Ack(m.Stream, group, m.ID)
	if err != nil {
		log.Warnf("error acknowledging message: %s", err)
	}
}
