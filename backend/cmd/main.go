package main

import (
	"WB/internal/config"
	"WB/internal/delivery/handlers"
	mwLogger "WB/internal/delivery/middleware/logger"
	kafka "WB/internal/lib/kafka"
	"WB/internal/lib/logger/sl"
	"WB/internal/lib/logger/slogpretty"
	"WB/internal/repository/postgres"
	"WB/internal/repository/redis"
	usecase "WB/internal/usecase"
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"golang.org/x/sync/errgroup"
)

func main() {
	//load cfg
	cfg := config.MustLoad()

	//init loger
	log := slogpretty.SetupLogger(cfg.Env)
	log.Info("starting server", slog.String("env", cfg.Env))

	connStr := "user=user password=password dbname=mydatabase sslmode=disable host=localhost port=5432"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	orderRepo, err := postgres.New(db, cfg.MigrationsPath)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	redisConn, err := redis.New(cfg)
	if err != nil {
		log.Error("failed to init redis", sl.Err(err))
		os.Exit(1)
	}

	kafkaProducer, err := kafka.NewProducer(cfg.Kafka.Brokers, cfg.Kafka.Topic)
	if err != nil {
		log.Error("Failed to initialize Kafka producer:", sl.Err(err))
		os.Exit(1)
	}

	orderUseCase := usecase.NewOrderUseCase(orderRepo, redisConn, kafkaProducer)

	kafkaConsumer := kafka.NewConsumer(cfg.Kafka.Brokers, cfg.Kafka.ConsumerGroup, cfg.Kafka.Topic)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// ErrGroup для управления всеми горутинами
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error{
		log.Info("Запуск Kafka consumer...")
		return kafkaConsumer.Start(context.Background(), orderUseCase.HandleMessage)
	})

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	router.Post("/order/new", handlers.NewOrder(log, orderUseCase))
	router.Get("/order/{id}", handlers.GetOrder(log, orderUseCase))

	log.Info("starting server", slog.String("address", cfg.Address))

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	g.Go(func() error {
		log.Info("starting HTTP server", slog.String("address", cfg.Address))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("HTTP server error", sl.Err(err))
			return err
		}
		return nil
	})
	<-ctx.Done()
	log.Info("shutting down gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("HTTP server forced shutdown", sl.Err(err))
	}

	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		log.Error("error during shutdown", sl.Err(err))
	}

	log.Info("closing resources...")
	if err := db.Close(); err != nil {
		log.Error("error closing database", sl.Err(err))
	}
	if err := redisConn.Close(); err != nil {
		log.Error("error closing redis", sl.Err(err))
	}
	if err := kafkaProducer.Close(); err != nil {
		log.Error("error closing kafka producer", sl.Err(err))
	}

	log.Info("server stopped gracefully")
}
