package main

import (
	"context"
	"github.com/gin-gonic/gin"

	"github.com/kind84/gospiga/api"
	"github.com/kind84/gospiga/db/dgraph"
	"github.com/kind84/gospiga/domain"
	"github.com/kind84/gospiga/pkg/redis"
	"github.com/kind84/gospiga/streamer"
	"github.com/kind84/gospiga/usecase"
)

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
	app := usecase.NewApp(ctx, ds, db, streamer)
	service := api.NewService(app)

	r := gin.Default()
	r.GET("/ping", service.Ping)
	r.POST("/new-recipe", service.NewRecipe)
	r.Run()
}
