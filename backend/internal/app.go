package internal

import (
	"WB/internal/model"
	"WB/internal/service/kafka"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/redis/go-redis/v9"

	// "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	HttpServer    *fiber.App
	Database      *pgxpool.Pool
	Redis         *redis.Client
	KafkaProducer *kafka.Producer
	KafkaConsumer *kafka.Consumer
}

func (a *App) Init(ctx context.Context) {
	a.SetupDb()
	a.SetupRedis()
	a.SetupKafka()
	a.setupHttp()

	go a.runConsumer(ctx)

	// log.Fatal(a.HttpServer.Listen(":3000"))
}

func (a *App) SetupDb() {
	// Строка подключения к базе данных
	connStr := "postgresql://user:password@localhost:5432/mydatabase"

	// Инициализация пула подключений
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatalf("не удалось подключиться к базе данных: %v", err)
	}
	a.Database = pool

	// Чтение SQL-схемы из файла
	sqlBytes, err := os.ReadFile("./pkg/database/schema.sql")
	if err != nil {
		log.Fatalf("не удалось прочитать файл схемы: %v", err)
	}
	createTablesSQL := string(sqlBytes)

	// Выполнение SQL для создания таблиц
	commandTag, err := a.Database.Exec(context.Background(), createTablesSQL)
	if err != nil {
		log.Fatalf("не удалось выполнить схему: %v", err)
	}

	// Логирование результата
	log.Printf("Схема базы данных успешно инициализирована. Команда: %s, Затронуто строк: %d",
		commandTag.String(), commandTag.RowsAffected())
}

func (a *App) SetupRedis() {
	a.Redis = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	statusCmd := a.Redis.Ping(context.Background())
	log.Println(statusCmd.String())
	log.Println("Redis initialized successfully")
}

func (a *App) SetupKafka() {
	producer, err := kafka.NewProducer([]string{"localhost:9092"}, "orders")
	if err != nil {
		log.Fatalf("Failed to initialize Kafka producer: %v", err)
	}
	a.KafkaProducer = producer
	log.Println("Kafka producer initialized successfully")

	consumer, err := kafka.NewConsumer([]string{"localhost:9092"}, "orders-group", "orders")
	if err != nil {
		log.Fatalf("Failed to initialize Kafka consumer: %v", err)
	}
	a.KafkaConsumer = consumer
	log.Println("Kafka consumer initialized successfully")
}

func (a *App) setupHttp() {
	app := fiber.New()
	app.Use(cors.New())

	app.Get("/", func(c fiber.Ctx) error {
		log.Println("Получен GET запрос на /")
		return c.SendFile("./static/index.html")
	})

	app.Get("/api/orders/:order_uid", func(c fiber.Ctx) error {
		orderUID := c.Params("order_uid")

		orderJSON, err := a.Redis.Get(context.Background(), orderUID).Result()
		if err == nil {
			log.Printf("Заказ %s найден в Redis", orderUID)
			return c.Status(fiber.StatusOK).SendString(orderJSON)
		}
		if err != redis.Nil {
			log.Printf("Ошибка чтения из Redis для заказа %s: %v", orderUID, err)
			// Продолжаем с PostgreSQL
		} else {
			log.Printf("Заказ %s не найден в Redis, проверяем PostgreSQL", orderUID)
		}

		// Fetch from PostgreSQL
		order, err := a.getOrderFromPostgres(orderUID)
		if err != nil {
			log.Printf("Заказ %s не найден в PostgreSQL", orderUID)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to fetch items",
			})
		}

		log.Printf("Успешно получен заказ %s из PostgreSQL", orderUID)
		return c.Status(fiber.StatusOK).JSON(order)
	})

	app.Post("/api/create_order", func(c fiber.Ctx) error {
		var order model.Order

		// Парсинг JSON body в структуру Order
		if err := c.Bind().Body(&order); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid JSON format",
			})
		}

		// Validate required fields
		if order.OrderUID == "" || order.TrackNumber == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "OrderUID and TrackNumber are required",
			})
		}

		// Конвертация в JSON для Kafka
		orderJSON, err := json.Marshal(order)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to process order",
			})
		}

		// Сообщение в кафку
		err = a.KafkaProducer.Send(context.Background(), order.OrderUID, orderJSON)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to send order to Kafka",
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message":   "Order successfully sent to Kafka",
			"order_uid": order.OrderUID,
		})
	})

	a.HttpServer = app
}

func (a *App) getOrderFromPostgres(orderUID string) (model.Order, error) {
	var order model.Order
	err := a.Database.QueryRow(context.Background(), `
		SELECT order_uid, track_number, entry, payment_transaction, locale, internal_signature, customer_id, 
				delivery_service, shardkey, sm_id, date_created, oof_shard
		FROM orders WHERE order_uid = $1`, orderUID).Scan(
		&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Payment.Transaction, &order.Locale,
		&order.InternalSignature, &order.CustomerID, &order.DeliveryService,
		&order.Shardkey, &order.SmID, &order.DateCreated, &order.OofShard)
	if err != nil {
		log.Printf("Ошибка чтения заказа %s из orders: %v", orderUID, err)
		return model.Order{}, err
	}

	err = a.Database.QueryRow(context.Background(), `
		SELECT name, phone, zip, city, address, region, email
		FROM delivery WHERE order_uid = $1`, orderUID).Scan(
		&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip,
		&order.Delivery.City, &order.Delivery.Address, &order.Delivery.Region,
		&order.Delivery.Email)
	if err != nil {
		log.Printf("Ошибка чтения delivery для заказа %s: %v", orderUID, err)
		return model.Order{}, err
	}

	err = a.Database.QueryRow(context.Background(), `
		SELECT request_id, currency, provider, amount, payment_dt, 
				bank, delivery_cost, goods_total, custom_fee
		FROM payment WHERE transaction = $1`, order.Payment.Transaction).Scan(
		&order.Payment.RequestID, &order.Payment.Currency,
		&order.Payment.Provider, &order.Payment.Amount, &order.Payment.PaymentDt,
		&order.Payment.Bank, &order.Payment.DeliveryCost, &order.Payment.GoodsTotal,
		&order.Payment.CustomFee)
	if err != nil {
		log.Printf("Ошибка чтения payment для заказа %s: %v", orderUID, err)
		return model.Order{}, err
	}

	rows, err := a.Database.Query(context.Background(), `
		SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
		FROM items WHERE order_uid = $1`, orderUID)
	if err != nil {
		log.Printf("Ошибка чтения items для заказа %s: %v", orderUID, err)
		return model.Order{}, err
	}
	defer rows.Close()

	order.Items = []model.Item{}
	for rows.Next() {
		var item model.Item
		if err := rows.Scan(&item.ChrtID, &item.TrackNumber, &item.Price, &item.Rid,
			&item.Name, &item.Sale, &item.Size, &item.TotalPrice,
			&item.NmID, &item.Brand, &item.Status); err != nil {
			log.Printf("Ошибка сканирования items для заказа %s: %v", orderUID, err)
			return model.Order{}, err
		}
		order.Items = append(order.Items, item)
	}

	log.Printf("Успешно получен заказ %s из PostgreSQL", orderUID)
	return order, nil
}

func (a *App) runConsumer(ctx context.Context) error {
	const maxRetries = 3
	const retryDelay = 5 * time.Second

	for {
		select {
		case <-ctx.Done():
			log.Println("Получен сигнал завершения, остановка консьюмера")
			a.KafkaConsumer.Close()
			return nil
		default:
			// Чтение сообщения из Kafka
			key, value, err := a.KafkaConsumer.Consume(ctx)
			
			if err != nil {
				log.Printf("Ошибка при получении сообщения из Kafka: %v, остановка консьюмера", err)
				a.KafkaConsumer.Close()
				return nil
			}

			// Парсинг JSON
			var order model.Order
			if err := json.Unmarshal(value, &order); err != nil {
				log.Printf("Ошибка парсинга JSON заказа %s: %v", string(key), err)
				continue
			}

			// Обработка заказа с ретраями
			for attempt := 1; attempt <= maxRetries; attempt++ {
				err = a.processOrder(ctx, &order)
				if err == nil {
					log.Printf("Заказ %s успешно обработан", order.OrderUID)
					break
				}

				log.Printf("Ошибка обработки заказа %s (попытка %d/%d): %v", order.OrderUID, attempt, maxRetries, err)
				if attempt < maxRetries {
					time.Sleep(retryDelay)
				} else {
					log.Printf("Исчерпаны попытки обработки заказа %s, пропускаем", order.OrderUID)
				}
			}
		}
	}
}

func (a *App) processOrder(ctx context.Context, order *model.Order) error {
	// Вставка в Postgres
	if err := a.insertOrderToPostgres(ctx, order); err != nil {
		return fmt.Errorf("ошибка вставки в Postgres: %w", err)
	}

	// Кэширование в Redis (не критично, поэтому не возвращаем ошибку)
	if err := a.cacheOrderToRedis(ctx, order); err != nil {
		log.Printf("Ошибка кэширования заказа %s в Redis: %v", order.OrderUID, err)
	}

	return nil
}

func (a *App) insertOrderToPostgres(ctx context.Context, order *model.Order) error {
	log.Printf("Вставка заказа %s в PostgreSQL", order.OrderUID)
	tx, err := a.Database.Begin(ctx)
	if err != nil {
		return fmt.Errorf("ошибка начала транзакции: %w", err)
	}
	defer tx.Rollback(ctx)

	log.Println("Вставка в таблицу delivery")
	result, err := tx.Exec(ctx, `
        INSERT INTO delivery (order_uid, name, phone, zip, city, address, region, email)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        ON CONFLICT (order_uid) DO NOTHING`,
		order.OrderUID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip,
		order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email)
	if err != nil {
		return fmt.Errorf("ошибка вставки в delivery: %w", err)
	}
	rowsAffected := result.RowsAffected()
	log.Printf("Вставлено строк в delivery: %d", rowsAffected)

	log.Println("Вставка в таблицу payment")
	result, err = tx.Exec(ctx, `
        INSERT INTO payment (transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
        ON CONFLICT (transaction) DO NOTHING`,
		order.Payment.Transaction, order.Payment.RequestID, order.Payment.Currency, order.Payment.Provider,
		order.Payment.Amount, order.Payment.PaymentDt, order.Payment.Bank, order.Payment.DeliveryCost,
		order.Payment.GoodsTotal, order.Payment.CustomFee)
	if err != nil {
		return fmt.Errorf("ошибка вставки в payment: %w", err)
	}
	rowsAffected = result.RowsAffected()
	log.Printf("Вставлено строк в payment: %d", rowsAffected)

	log.Println("Вставка в таблицу orders")
	result, err = tx.Exec(ctx, `
        INSERT INTO orders (order_uid, track_number, entry, delivery_uid, payment_transaction, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
        ON CONFLICT (order_uid) DO NOTHING`,
		order.OrderUID, order.TrackNumber, order.Entry, order.OrderUID, order.Payment.Transaction,
		order.Locale, order.InternalSignature, order.CustomerID, order.DeliveryService,
		order.Shardkey, order.SmID, order.DateCreated, order.OofShard)
	if err != nil {
		return fmt.Errorf("ошибка вставки в orders: %w", err)
	}
	rowsAffected = result.RowsAffected()
	log.Printf("Вставлено строк в orders: %d", rowsAffected)

	log.Println("Вставка в таблицу items")
	for i, item := range order.Items {
		result, err = tx.Exec(ctx, `
            INSERT INTO items (order_uid, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
			order.OrderUID, item.ChrtID, item.TrackNumber, item.Price, item.Rid, item.Name,
			item.Sale, item.Size, item.TotalPrice, item.NmID, item.Brand, item.Status)
		if err != nil {
			return fmt.Errorf("ошибка вставки в items (элемент %d): %w", i, err)
		}
		rowsAffected = result.RowsAffected()
		log.Printf("Вставлено строк в items (элемент %d): %d", i, rowsAffected)
	}

	log.Println("Коммит транзакции")
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("ошибка коммита транзакции: %w", err)
	}
	log.Println("Транзакция успешно закоммичена")
	return nil
}

func (a *App) cacheOrderToRedis(ctx context.Context, order *model.Order) error {
	orderJSON, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("ошибка сериализации заказа: %w", err)
	}

	key := order.OrderUID
	return a.Redis.Set(ctx, key, orderJSON, 24*time.Hour).Err()
}
