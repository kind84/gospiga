package main

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/kind84/gospiga/finder/api"
	"github.com/kind84/gospiga/finder/db"
	"github.com/kind84/gospiga/finder/fulltext"
	"github.com/kind84/gospiga/finder/usecase"
	"github.com/kind84/gospiga/pkg/redis"
	"github.com/kind84/gospiga/pkg/streamer"
)

func init() {
	log.Info("Setting up configuration...")
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

	db := db.NewRedisDB(rdb)
	streamer := streamer.NewRedisStreamer(rdb)

	app := usecase.NewApp(ctx, db, ft, streamer)
	service := api.NewService(app)

	r := gin.Default()
	r.POST("/search-recipe", service.SearchRecipe)
	r.Run()
}
