package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"github.com/kind84/gospiga/api"
	"github.com/kind84/gospiga/db/dgraph"
	"github.com/kind84/gospiga/domain"
	"github.com/kind84/gospiga/pkg/redis"
	"github.com/kind84/gospiga/provider"
	"github.com/kind84/gospiga/streamer"
	"github.com/kind84/gospiga/usecase"
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

	rdb, err := redis.NewClient("localhost:6379")
	if err != nil {
		panic("can't connect to redis")
	}

	db, err := dgraph.NewDB()
	if err != nil {
		panic("can't connect to database")
	}

	ds := domain.NewService(db)
	streamer := streamer.NewRedisStreamer(rdb)

	token := viper.GetString("dato.token")
	if token == "" {
		panic("missing dato cms token")
	}
	provider, err := provider.NewDatoProvider(token)
	if err != nil {
		panic("can't connect to dato cms")
	}

	app := usecase.NewApp(ctx, ds, db, streamer, provider)
	service := api.NewService(app)

	r := gin.Default()
	r.GET("/ping", service.Ping)
	r.POST("/new-recipe", service.NewRecipe)
	r.Run()
}
