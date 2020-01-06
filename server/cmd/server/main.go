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
	"github.com/kind84/gospiga/server/api"
	"github.com/kind84/gospiga/server/db/dgraph"
	"github.com/kind84/gospiga/server/domain"
	"github.com/kind84/gospiga/server/provider"
	"github.com/kind84/gospiga/server/usecase"
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

	db, err := dgraph.NewDB(ctx)
	if err != nil {
		log.Fatalf("can't connect to database: %s", err)
	}

	ds := domain.NewService(db)
	streamer := streamer.NewRedisStreamer(rdb)

	token := viper.GetString("dato.token")
	if token == "" {
		panic("missing dato cms token")
	}
	provider, err := provider.NewDatoProvider(token)
	if err != nil {
		log.Fatalf("can't connect to dato cms: %s", err)
	}

	app := usecase.NewApp(ctx, ds, db, streamer, provider)
	service := api.NewService(app)

	r := gin.Default()
	r.GET("/ping", service.Ping)
	r.POST("/new-recipe", service.NewRecipe)
	r.Run()
}
