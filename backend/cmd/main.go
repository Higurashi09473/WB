package main

import (
	"WB/internal"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := internal.App{}
	app.Init(ctx)

	go func() {
		if err := app.HttpServer.Listen(":3000"); err != nil {
			log.Printf("Ошибка HTTP-сервера: %v", err)
			cancel()
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Goroutine to handle shutdown signals
	go func() {
		sig := <-sigChan
		log.Printf("Received signal: %s, initiating shutdown...", sig)

		cancel()

		// Shutdown HTTP server with a timeout
		ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
		defer cancel() 

		if err := app.HttpServer.ShutdownWithContext(ctx); err != nil {
			log.Printf("HTTP server shutdown error: %v", err)
		} else {
			log.Println("HTTP server shutdown successfully")
		}

		// Close database connection
		if app.Database != nil {
			app.Database.Close()
			log.Println("Database connection closed")
		}

		// Close Redis connection
		if app.Redis != nil {
			if err := app.Redis.Close(); err != nil {
				log.Printf("Redis connection close error: %v", err)
			} else {
				log.Println("Redis connection closed")
			}
		}

		// Close Kafka producer and consumer
		if app.KafkaProducer != nil {
			app.KafkaProducer.Close()
			log.Println("Kafka producer closed")
		}
		// if app.KafkaConsumer != nil {
		// 	app.KafkaConsumer.Close()
		// 	log.Println("Kafka consumer closed")
		// }
		log.Println("Application shutdown complete")
		os.Exit(0)
	}()

	// Keep the main goroutine running
	select {}
}