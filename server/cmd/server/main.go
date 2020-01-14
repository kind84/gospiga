package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"google.golang.org/grpc"

	"github.com/kind84/gospiga/pkg/redis"
	"github.com/kind84/gospiga/pkg/streamer"
	pb "github.com/kind84/gospiga/proto"
	"github.com/kind84/gospiga/server/api"
	"github.com/kind84/gospiga/server/db/dgraph"
	"github.com/kind84/gospiga/server/domain"
	gogrpc "github.com/kind84/gospiga/server/grpc"
	"github.com/kind84/gospiga/server/provider"
	"github.com/kind84/gospiga/server/usecase"
)

const defaultFinderPort = "50051"

func init() {
	log.SetLevel(log.DebugLevel)

	log.Info("Setting up configuration...")
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	// viper.SetEnvPrefix("gospiga")
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
		log.Fatal("missing dato cms token")
	}
	provider, err := provider.NewDatoProvider(token)
	if err != nil {
		log.Fatalf("can't connect to dato cms: %s", err)
	}

	finderPort := viper.GetString("FINDER_PORT")
	if finderPort == "" {
		finderPort = defaultFinderPort
	}
	conn, err := grpc.Dial(fmt.Sprintf("finder:%s", finderPort))
	if err != nil {
		log.Fatalf("can't connect to finder grpc server: %s", err)
	}
	defer conn.Close()

	grpcClient := pb.NewFinderClient(conn)
	stub := gogrpc.NewStub(&grpcClient)

	app := usecase.NewApp(ctx, ds, db, streamer, provider, stub)
	service := api.NewService(app)

	r := gin.Default()
	r.GET("/ping", service.Ping)
	r.POST("/new-recipe", service.NewRecipe)
	r.POST("/search-recipes", service.SearchRecipes)
	r.Run()
}
