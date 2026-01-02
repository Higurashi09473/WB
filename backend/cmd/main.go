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

	chiprom "github.com/766b/chi-prometheus" // или "github.com/yarlson/chiprom"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/sync/errgroup"
)

func main() {
	cfg := config.MustLoad()

	log := slogpretty.SetupLogger(cfg.Env)
	log.Info("starting server", slog.String("env", cfg.Env))

	db, err := sql.Open("pgx", cfg.Postgresql.DSN())
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	orderRepo := postgres.MustLoad(db, cfg.MigrationsPath)

	redisConn := redis.MustLoad(cfg.Redis.Host, cfg.Redis.Port, cfg.Redis.Password, cfg.Redis.DB)

	kafkaProducer := kafka.MustProducer(cfg.Kafka.Brokers, cfg.Kafka.Topic)

	orderUseCase := usecase.NewOrderUseCase(orderRepo, redisConn, kafkaProducer)

	kafkaConsumer := kafka.NewConsumer(cfg.Kafka.Brokers, cfg.Kafka.ConsumerGroup, cfg.Kafka.Topic, cfg.Kafka.DLQTopic)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		log.Info("starting Kafka consumer...")
		return kafkaConsumer.Start(ctx, orderUseCase.HandleMessage)
	})

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	
	// позже нужно добавить метрики
	router.Use(chiprom.NewMiddleware("my-service"))
	// router.Use(middleware.URLFormat)

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://0.0.0.0:*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	router.Handle("/metrics", promhttp.Handler())
	router.Post("/api/create_order", handlers.NewOrder(log, orderUseCase))
	router.Get("/api/orders/{id}", handlers.GetOrder(log, orderUseCase))
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/static/index.html")
	})

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	g.Go(func() error {
		log.Info("starting HTTP server", slog.String("address", cfg.Address))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
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
