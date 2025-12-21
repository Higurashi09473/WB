package main

import (
	"WB/internal/config"
	"WB/internal/delivery/handlers"
	mwLogger "WB/internal/delivery/middleware/logger"
	"WB/internal/lib/logger/sl"
	"WB/internal/lib/logger/slogpretty"
	"WB/internal/repository/postgres"
	"WB/internal/repository/redis"
	"database/sql"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	//load cfg
	cfg := config.MustLoad()

	//init loger
	log := setupLogger(cfg.Env)
	log.Info("starting server", slog.String("env", cfg.Env))

	connStr := "user=user password=password dbname=mydatabase sslmode=disable host=localhost port=5432"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	//init postgres
	storage, err := postgres.New(db, cfg.MigrationsPath)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}
	defer func() {
		if err := storage.Close(); err != nil {
			log.Error("failed to init redis", sl.Err(err))
			os.Exit(1)
		}
	}()


	//инит редис
	redisConn, err := redis.New(cfg)
	if err != nil{
		log.Error("failed to init redis", sl.Err(err))
		os.Exit(1)
	}
	defer func() {
		if err := redisConn.Close(); err != nil {
			log.Error("failed to close redis", sl.Err(err))
		}
	}()

	//запуск консьюмера

	//запуск продюссера
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


	router.Post("/order/new", handlers.NewOrder(log, storage))
	router.Get("/order/{id}", handlers.GetOrder(log, storage))


	log.Info("starting server", slog.String("address", cfg.Address))

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Error(err.Error())

		log.Error("failed to start server")
	}

	log.Error("server stopped")

	// грейсфулл шотдаун
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
