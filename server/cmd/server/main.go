package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"google.golang.org/grpc"

	"gospiga"
	"gospiga/pkg/log"
	"gospiga/pkg/provider"
	"gospiga/pkg/redis"
	"gospiga/pkg/streamer"
	pb "gospiga/proto"
	"gospiga/server/api"
	"gospiga/server/db/dgraph"
	"gospiga/server/domain"
	gogrpc "gospiga/server/grpc"
	"gospiga/server/usecase"
)

const defaultFinderPort = "50051"

func init() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	// viper.SetEnvPrefix("gospiga")
	viper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.ReadInConfig()
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("panic trapped in main goroutine : %+v", err)
			log.Errorf("stacktrace from panic: %s", string(debug.Stack()))
			os.Exit(1)
		}
	}()

	gospiga.PrintVersion(os.Stdout)

	ctx := context.Background()
	var stop context.CancelFunc
	// Wire shutdownCh to get events depending on the OS we are running in
	if runtime.GOOS == "windows" {
		fmt.Println("Listening to Windows OS interrupt signal for graceful shutdown.")
		ctx, stop = signal.NotifyContext(ctx, os.Interrupt)

	} else {
		fmt.Println("Listening to SIGINT or SIGTERM for graceful shutdown.")
		ctx, stop = signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	}
	defer stop()

	rdb, err := redis.NewClient("redis:6379")
	if err != nil {
		log.Fatalf("can't connect to redis: %s", err)
	}

	db, err := dgraph.NewDB(ctx)
	if err != nil {
		log.Fatalf("error initializing database: %s", err)
	}

	ds := domain.NewService(db)
	streamer, err := streamer.NewRedisStreamer(rdb)
	if err != nil {
		log.Fatalf("error initializing redis streamer: %s", err)
	}

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
	conn, err := grpc.Dial(fmt.Sprintf("finder:%s", finderPort), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("can't connect to finder grpc server: %s", err)
	}
	defer conn.Close()

	grpcClient := pb.NewFinderClient(conn)
	stub := gogrpc.NewStub(&grpcClient)

	app := usecase.NewApp(ctx, ds, db, streamer, provider, stub)
	service := api.NewService(app)

	config := cors.DefaultConfig()
	config.AddAllowHeaders("X-Apollo-Tracing")
	config.AllowAllOrigins = true
	c := cors.New(config)

	r := gin.Default()
	r.Use(c)
	r.LoadHTMLFiles("/templates/graphql-playground.html")
	g := r.Group("/server")
	{
		g.Group("/server")
		g.GET("/ping", service.Ping)
		g.GET("/x/gql/play", func(c *gin.Context) {
			c.HTML(http.StatusOK, "graphql-playground.html", gin.H{
				"title": "GraphQL Playground",
			})
		})
		g.POST("/new-recipe", service.NewRecipe)
		g.POST("/updated-recipe", service.UpdatedRecipe)
		g.POST("/deleted-recipe", service.DeletedRecipe)
		g.POST("/all-tags-images", service.AllTagsImages)
		g.POST("/load-recipes", service.LoadRecipes)
	}
	go r.Run()

	// wait for shutdown
	<-ctx.Done()
	stop()
	fmt.Println("\nShutdown signal detected, gracefully shutting down...")
	time.Sleep(time.Second)
	fmt.Println("bye")
}
