package main

import (
	"WB/internal"
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	defer cancel()

	app := internal.App{}
	app.Init(ctx, wg)

	go func() {
		if err := app.HttpServer.Listen(":3000"); err != nil {
			log.Printf("Ошибка HTTP-сервера: %v", err)
			cancel()
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("Received signal: %s, initiating shutdown...", sig)

		cancel()


		ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
		defer cancel() 

		if err := app.HttpServer.ShutdownWithContext(ctx); err != nil {
			log.Printf("HTTP server shutdown error: %v", err)
		} else {
			log.Println("HTTP server shutdown successfully")
		}

		if app.Database != nil {
			app.Database.Close()
			log.Println("Database connection closed")
		}

		if app.Redis != nil {
			if err := app.Redis.Close(); err != nil {
				log.Printf("Redis connection close error: %v", err)
			} else {
				log.Println("Redis connection closed")
			}
		}

		if app.KafkaProducer != nil {
			app.KafkaProducer.Close()
			log.Println("Kafka producer closed")
		}
		wg.Wait()
		log.Println("Application shutdown complete")
		os.Exit(0)
	}()
	select {}
}
