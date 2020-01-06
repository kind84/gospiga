package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"github.com/kind84/gospiga/pkg/redis"
	"github.com/kind84/gospiga/pkg/streamer"
	"github.com/kind84/gospiga/searcher/api"
	"github.com/kind84/gospiga/searcher/domain"
	"github.com/kind84/gospiga/searcher/fulltext"
	"github.com/kind84/gospiga/searcher/usecase"
)

func init() {
	fmt.Println("Setting up configuration...")
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.SetEnvPrefix("gospiga")
	viper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.ReadInConfig()
}

func main() {
	ctx := context.Background()

	rdb, err := redis.NewClient("redis:6379")
	if err != nil {
		log.Fatalf("can't connect to redis: %s", err)
	}

	ft := fulltext.NewRedisFT("redis:6379")
	if ft == nil {
		log.Fatal("cannot initialize redis fulltext")
	}

	ds := domain.NewService(rdb)
	streamer := streamer.NewRedisStreamer(rdb)

	app := usecase.NewApp(ctx, ft, ds, streamer)
	service := api.NewService(app)

	r := gin.Default()
	r.POST("/search-recipe", service.SearchRecipe)
	r.Run()
}
