package main

import (
	"context"
	"fmt"
	"net"
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
	"gospiga/finder/api"
	"gospiga/finder/db"
	"gospiga/finder/fulltext"
	gogrpc "gospiga/finder/grpc"
	"gospiga/finder/usecase"
	"gospiga/pkg/log"
	"gospiga/pkg/redis"
	"gospiga/pkg/streamer"
	pb "gospiga/proto"
)

const defaultPort = "50051"

func init() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.SetEnvPrefix("gospiga")
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

	ft, err := fulltext.NewRedisFT("redis:6379")
	if err != nil {
		log.Fatalf("cannot initialize redis fulltext: %s", err)
	}

	db := db.NewRedisDB(rdb)
	streamer, err := streamer.NewRedisStreamer(rdb)
	if err != nil {
		log.Fatalf("error initializing redis streamer: %s", err)
	}

	app := usecase.NewApp(ctx, db, ft, streamer)
	if err != nil {
		log.Fatalf("cannot initalize application: %s", err)
	}
	service := api.NewService(app)

	server := gogrpc.NewFinderServer(app)
	grpcServer := grpc.NewServer()
	pb.RegisterFinderServer(grpcServer, server)

	port := viper.GetString("TCP_PORT")
	if port == "" {
		port = defaultPort
	}
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	go grpcServer.Serve(lis)

	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	c := cors.New(config)

	r := gin.Default()
	r.Use(c)
	g := r.Group("/finder")
	{
		g.POST("/search-recipes", service.SearchRecipes)
		g.POST("/search-by-tag", service.SearchByTag)
		g.POST("/all-recipe-tags", service.AllRecipeTags)
	}
	go r.Run()

	// wait for shutdown
	<-ctx.Done()
	stop()
	fmt.Println("\nShutdown signal detected, gracefully shutting down...")
	time.Sleep(time.Second)
	fmt.Println("bye")
}
