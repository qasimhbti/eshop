package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eshop/config"
	"github.com/eshop/pkg/httphealthcheck"
	"github.com/eshop/pkg/utils"
	"github.com/eshop/version"
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
)

const healthCheckAddr = "/health-check"

func main() {
	err := run()
	if err != nil {
		handleError(err)
	}
}

func run() error {
	utils.InitLog()
	config, err := config.GetConfigs()
	if err != nil {
		return errors.WithMessage(err, "get config")
	}
	utils.LogStart(version.Version, config.Env)

	// DB Connection...
	DBClient, err := newMongoDBClientGetter(config.DBConString)
	if err != nil {
		return errors.WithMessage(err, "mongo client")
	}
	defer DBClient.Disconnect(context.Background())

	db := getmgoDB(DBClient, config.DBName)

	// Redis Connection...
	redisClient := redis.NewClient(&redis.Options{
		Addr: config.RedisConString,
	})

	_, err = redisClient.Ping().Result()
	if err != nil {
		return errors.WithMessage(err, "redis client")
	}
	log.Println("Redis ping successfully")

	httpServer := startHTTPServer(config, db, redisClient)
	if err != nil {
		return errors.WithMessage(err, "start HTTP Server")
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errChan := make(chan error)
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer signal.Stop(stopChan)

	// HTTP Server Go Routine
	go func() {
		log.Printf("HTTP server is up and running on %s port.", config.HTTPPort)
		if err := httpServer.ListenAndServe(); err != nil {
			if err.Error() != "http: Server closed" {
				errChan <- err
			}
		}
	}()

	// HTTP Health Check Server
	go httphealthcheck.Check(healthCheckAddr, errChan)

	defer func() {
		log.Println("Gracefully Shutting Down HTTP Server...")
		time.Sleep(5 * time.Second)
		httpServer.Shutdown(ctx)
		close(errChan)
		close(stopChan)
	}()

	select {
	case err := <-errChan:
		log.Printf("fatal error HTTP Server: %v\n", err)
	case <-stopChan:
		log.Println("receive shutdown/terminal signal")
	case <-ctx.Done():
		cancel()
	}

	return nil

}

func handleError(err error) {
	log.Fatalf("%s", err)
}
