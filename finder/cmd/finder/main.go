package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"google.golang.org/grpc"

	"github.com/kind84/gospiga/finder/api"
	"github.com/kind84/gospiga/finder/db"
	"github.com/kind84/gospiga/finder/fulltext"
	gogrpc "github.com/kind84/gospiga/finder/grpc"
	"github.com/kind84/gospiga/finder/usecase"
	"github.com/kind84/gospiga/pkg/redis"
	"github.com/kind84/gospiga/pkg/streamer"
	pb "github.com/kind84/gospiga/proto"
)

const defaultPort = "50051"

func init() {
	log.SetLevel(log.DebugLevel)

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
	shutdownCh := make(chan os.Signal, 1)

	// Wire shutdownCh to get events depending on the OS we are running in
	if runtime.GOOS == "windows" {
		fmt.Println("Listening to Windows OS interrupt signal for graceful shutdown.")
		signal.Notify(shutdownCh, os.Interrupt)

	} else {
		fmt.Println("Listening to SIGINT or SIGTERM for graceful shutdown.")
		signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM)
	}

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

	app := usecase.NewApp(db, ft, streamer)
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
	grpcServer.Serve(lis)

	r := gin.Default()
	r.POST("/search-recipe", service.SearchRecipes)
	go r.Run()

	// wait for shutdown
	if <-shutdownCh != nil {
		fmt.Println("\nShutdown signal detected, gracefully shutting down...")
		app.CloseGracefully()
	}
	fmt.Println("bye")
}
