package internal

import (
	"WB/internal/service"
	"context"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"

	// "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	httpServer    *fiber.App
	database      *pgxpool.Pool
	redis         *redis.Client
	kafkaProducer *service.Producer
	kafkaConsumer *service.Consumer
}

func (a *App) Init() {
	a.setupDb()
	a.SetupRedis()
	a.SetupKafka()
	a.setupHttp()

	log.Fatal(a.httpServer.Listen(":3000"))
}

func (a *App) setupDb() {
	connStr := "postgresql://user:password@localhost:5432/mydatabase"
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		panic(fmt.Errorf("failed to connect to database: %w", err))
	}
	a.database = pool

	// SQL to create tables if they don't exist
	createTablesSQL := `
		CREATE TABLE IF NOT EXISTS delivery (
		order_uid VARCHAR(255) PRIMARY KEY,
		name TEXT NOT NULL,
		phone TEXT NOT NULL,
		zip TEXT NOT NULL,
		city TEXT NOT NULL,
		address TEXT NOT NULL,
		region TEXT NOT NULL,
		email TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS payment (
		transaction VARCHAR(255) PRIMARY KEY,
		request_id TEXT,
		currency TEXT NOT NULL,
		provider TEXT NOT NULL,
		amount INTEGER NOT NULL,
		payment_dt BIGINT NOT NULL,
		bank TEXT NOT NULL,
		delivery_cost INTEGER NOT NULL,
		goods_total INTEGER NOT NULL,
		custom_fee INTEGER NOT NULL
	);

	CREATE TABLE IF NOT EXISTS orders (
		order_uid VARCHAR(255) PRIMARY KEY,
		track_number TEXT NOT NULL,
		entry TEXT NOT NULL,
		delivery_uid VARCHAR(255) NOT NULL,
		payment_transaction VARCHAR(255) NOT NULL,
		locale TEXT NOT NULL,
		internal_signature TEXT NOT NULL,
		customer_id TEXT NOT NULL,
		delivery_service TEXT NOT NULL,
		shardkey TEXT NOT NULL,
		sm_id INTEGER NOT NULL,
		date_created TIMESTAMP WITH TIME ZONE NOT NULL,
		oof_shard TEXT NOT NULL,
		FOREIGN KEY (delivery_uid) REFERENCES delivery(order_uid) ON DELETE CASCADE,
		FOREIGN KEY (payment_transaction) REFERENCES payment(transaction) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS items (
		id SERIAL PRIMARY KEY,
		order_uid VARCHAR(255) NOT NULL,
		chrt_id INTEGER NOT NULL,
		track_number TEXT NOT NULL,
		price INTEGER NOT NULL,
		rid TEXT NOT NULL,
		name TEXT NOT NULL,
		sale INTEGER NOT NULL,
		size TEXT NOT NULL,
		total_price INTEGER NOT NULL,
		nm_id INTEGER NOT NULL,
		brand TEXT NOT NULL,
		status INTEGER NOT NULL,
		FOREIGN KEY (order_uid) REFERENCES orders(order_uid) ON DELETE CASCADE
	);
    `

	// Execute the SQL to create tables
	commandTag, err := a.database.Exec(context.Background(), createTablesSQL)
	if err != nil {
		panic(fmt.Errorf("failed to create tables: %w", err))
	}

	// Вывод CommandTag
	log.Printf("CommandTag: %s, Rows Affected: %d", commandTag.String(), commandTag.RowsAffected())

	log.Println("Database tables initialized successfully")
}

func (a *App) SetupRedis() {
	a.redis = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	statusCmd := a.redis.Ping(context.Background())
	log.Println(statusCmd.String())
	log.Println("Redis initialized successfully")
}

func (a *App) SetupKafka() {
	// Initialize Kafka Producer
	producer, err := service.NewProducer([]string{"localhost:9092"}, "orders")
	if err != nil {
		log.Fatalf("Failed to initialize Kafka producer: %v", err)
	}
	a.kafkaProducer = producer
	log.Println("Kafka producer initialized successfully")

	// Initialize Kafka Consumer
	consumer, err := service.NewConsumer([]string{"localhost:9092"}, "orders-group", "orders")
	if err != nil {
		log.Fatalf("Failed to initialize Kafka consumer: %v", err)
	}
	a.kafkaConsumer = consumer
	log.Println("Kafka consumer initialized successfully")
}

func (a *App) setupHttp() {
	app := fiber.New()
	app.Use(cors.New())

	app.Get("/api", func(c fiber.Ctx) error {
		return c.SendString("Hello, World 👋!")
	})

	// app.Post("/api/orders", )

	a.httpServer = app
}

// func handlerCreateOrder(c fiber.Ctx) error {
// 	var order Order

// 	if err := c.BodyParser(&order); err != nil {
// 		return c.Status()
// 	}
// }
