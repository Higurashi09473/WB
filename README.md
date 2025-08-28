# Демонстрационный сервис с Kafka, PostgreSQL, кешем (L0)
```
Микросервис для управления заказами с интеграцией Kafka, PostgreSQL и Redis.
Принимает заказы через Kafka, сохраняет их в PostgreSQL, кэширует в Redis и предоставляет REST API для доступа к данным.
```
# Технологии

```
Go 1.24 — Язык программирования
Fiber — Высокопроизводительный HTTP-фреймворк
Goose - Инструмент для миграции базы данных
Kafka (segmentio/kafka-go) — брокер сообщений
PostgreSQL (pgx) — Хранилище данных
Redis (go-redis) — Кэширование заказов
Docker / Docker Compose — Контейнеризация
```

# Возможности

```
📥 Приём и валидация заказов через Kafka
💾 Хранение заказов, доставок, оплат и товаров в PostgreSQL
⚡ Кэширование заказов в Redis
🌐 REST API для создания и получения заказов
🖥 HTML-интерфейс для работы с заказами
```

# Быстрый старт

1. Запуск redis, kafka и postgresql
```
docker compose up -d
```

2. Запуск сервиса
```
go run ./cmd/app
```

# API
Создать заказ
```
Эндпоинт: POST /api/create_order
Описание: Отправляет заказ в Kafka для обработки
Пример:curl -X POST http://127.0.0.1:3000/api/create_order \
-H "Content-Type: application/json" \
-d '{
   "order_uid": "b563feb7b2b84b6test",
   "track_number": "WBILMTESTTRACK",
   "entry": "WBIL",
   "delivery": {
      "name": "Test Testov",
      "phone": "+9720000000",
      "zip": "2639809",
      "city": "Kiryat Mozkin",
      "address": "Ploshad Mira 15",
      "region": "Kraiot",
      "email": "test@gmail.com"
   },
   "payment": {
      "transaction": "b563feb7b2b84b6test",
      "request_id": "",
      "currency": "USD",
      "provider": "wbpay",
      "amount": 1817,
      "payment_dt": 1637907727,
      "bank": "alpha",
      "delivery_cost": 1500,
      "goods_total": 317,
      "custom_fee": 0
   },
   "items": [
      {
         "chrt_id": 9934930,
         "track_number": "WBILMTESTTRACK",
         "price": 453,
         "rid": "ab4219087a764ae0btest",
         "name": "Mascaras",
         "sale": 30,
         "size": "0",
         "total_price": 317,
         "nm_id": 2389212,
         "brand": "Vivienne Sabo",
         "status": 202
      }
   ],
   "locale": "en",
   "internal_signature": "",
   "customer_id": "test",
   "delivery_service": "meest",
   "shardkey": "9",
   "sm_id": 99,
   "date_created": "2021-11-26T06:22:19Z",
   "oof_shard": "1"
}'
```

Получить заказ

```
Эндпоинт: GET /api/orders/:order_uid
Описание: Возвращает заказ из Redis или PostgreSQL
Пример:curl http://localhost:3000/api/orders/b563feb7b2b84b6test
```



🖼 HTML-интерфейс

```
URL: http://localhost:3000/
Описание: Веб-интерфейс для создания и просмотра заказов
Файл: ./static/index.html
```

🔧 Команды Make
# Запустить приложение
```
make run
```

# Запустить Docker Compose
```
make docker-up
```

# Остановить Docker Compose
```
make docker-down
```

# 🛠 Структура проекта

```plaintext
WB
├── backend
│   ├── cmd
│   │   └── main.go
│   ├── deployments
│   │   ├── kafka-data
│   │   ├── postgres-data
│   │   ├── redis-data
│   │   └── docker-compose.yml
│   ├── internal
│   │   ├── app.go
│   │   ├── model
│   │   │   └── order.go
│   │   └── service
│   │       └── kafka
│   │           ├── consumer.go
│   │           └── producer.go
│   ├── pkg
│   │   └── database
│   │       ├── migrations
│   │       │   └── 20250828155143_create_initial_tables.sql
│   │       └── schema.sql
│   ├── static
│   │   └── index.html
│   ├── go.mod
│   ├── go.sum
│   └── Makefile
└── README.md
```
