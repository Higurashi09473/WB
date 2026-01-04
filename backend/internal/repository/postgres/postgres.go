// Package postgres provides PostgreSQL-based implementation of order storage.
// It handles database connection, migrations and order CRUD operations.
package postgres

import (
	"WB/internal/models"
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" //import pgx driver
	"github.com/pressly/goose/v3"
)

// Storage represents PostgreSQL repository for orders.
type Storage struct {
	db *sql.DB
}

// MustLoad initializes PostgreSQL storage with database connection and runs migrations.
// If a failure occurs during migration, os.exit is executed
func MustLoad(log *slog.Logger, db *sql.DB, migrationsPath string) *Storage {
	const op = "storage.postgres.MustLoad"

	db.SetMaxIdleConns(20)
	db.SetConnMaxLifetime(time.Hour)
	db.SetConnMaxIdleTime(15 * time.Minute)

	if err := db.Ping(); err != nil {
		log.Error("%s: db ping failed", op, slog.String("error", err.Error()))
		os.Exit(1)
	}

	if err := runMigrations(db, migrationsPath); err != nil {
		log.Error("%s: failed to apply database migrations", op, slog.String("error", err.Error()))
		os.Exit(1)
	}

	return &Storage{db: db}
}

// runMigrations applies pending migrations using goose.
func runMigrations(db *sql.DB, migrationsPath string) error {
	if migrationsPath == "" {
		return fmt.Errorf("migrations path is empty")
	}

	if err := goose.SetDialect("pgx"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Up(db, migrationsPath); err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}

// NewOrder adds a new order to the database or returns an error.
func (s *Storage) NewOrder(order models.Order) error {
	const op = "storage.postgres.NewOrder"

	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("%s: begin transaction: %w", op, err)
	}
	defer tx.Rollback()

	// 1. Delivery
	_, err = tx.Exec(`
        INSERT INTO delivery (order_uid, name, phone, zip, city, address, region, email)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        ON CONFLICT (order_uid) DO NOTHING`,
		order.OrderUID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip,
		order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email)
	if err != nil {
		return fmt.Errorf("%s: insert delivery: %w", op, err)
	}

	// 2. Payment
	_, err = tx.Exec(`
        INSERT INTO payment (transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
        ON CONFLICT (transaction) DO NOTHING`,
		order.Payment.Transaction, order.Payment.RequestID, order.Payment.Currency, order.Payment.Provider,
		order.Payment.Amount, order.Payment.PaymentDt, order.Payment.Bank, order.Payment.DeliveryCost,
		order.Payment.GoodsTotal, order.Payment.CustomFee)
	if err != nil {
		return fmt.Errorf("%s: insert payment: %w", op, err)
	}

	// 3. Orders
	_, err = tx.Exec(`
        INSERT INTO orders (order_uid, track_number, entry, delivery_uid, payment_transaction, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
        ON CONFLICT (order_uid) DO NOTHING`,
		order.OrderUID, order.TrackNumber, order.Entry, order.OrderUID, order.Payment.Transaction,
		order.Locale, order.InternalSignature, order.CustomerID, order.DeliveryService,
		order.Shardkey, order.SmID, order.DateCreated, order.OofShard)
	if err != nil {
		return fmt.Errorf("%s: insert orders: %w", op, err)
	}

	// 4. Items
	for _, item := range order.Items {
		_, err = tx.Exec(`
            INSERT INTO items (
                order_uid, chrt_id, track_number, price, rid, name, sale,
                size, total_price, nm_id, brand, status
            ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`, // уникальный ключ, предположительно
			order.OrderUID, item.ChrtID, item.TrackNumber, item.Price, item.Rid,
			item.Name, item.Sale, item.Size, item.TotalPrice, item.NmID, item.Brand, item.Status)
		if err != nil {
			return fmt.Errorf("%s: insert item %v: %w", op, item.ChrtID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%s: commit transaction: %w", op, err)
	}

	return nil
}

// GetOrder retrieves an order by its ID from the database.
func (s *Storage) GetOrder(orderID string) (models.Order, error) {
	const op = "storage.postgres.GetOrder"

	// 1. Orders
	var order models.Order
	err := s.db.QueryRow(`
		SELECT order_uid, track_number, entry, payment_transaction, locale, internal_signature, customer_id, 
				delivery_service, shardkey, sm_id, date_created, oof_shard
		FROM orders WHERE order_uid = $1`, orderID).Scan(
		&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Payment.Transaction, &order.Locale,
		&order.InternalSignature, &order.CustomerID, &order.DeliveryService,
		&order.Shardkey, &order.SmID, &order.DateCreated, &order.OofShard)
	if err != nil {
		return models.Order{}, fmt.Errorf("%s: get orders: %w", op, err)
	}

	// 2. Delivery
	err = s.db.QueryRow(`
		SELECT name, phone, zip, city, address, region, email
		FROM delivery WHERE order_uid = $1`, orderID).Scan(
		&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip,
		&order.Delivery.City, &order.Delivery.Address, &order.Delivery.Region,
		&order.Delivery.Email)
	if err != nil {
		return models.Order{}, fmt.Errorf("%s: get delivery: %w", op, err)
	}

	// 3. Payment
	err = s.db.QueryRow(`
		SELECT request_id, currency, provider, amount, payment_dt, 
				bank, delivery_cost, goods_total, custom_fee
		FROM payment WHERE transaction = $1`, order.Payment.Transaction).Scan(
		&order.Payment.RequestID, &order.Payment.Currency,
		&order.Payment.Provider, &order.Payment.Amount, &order.Payment.PaymentDt,
		&order.Payment.Bank, &order.Payment.DeliveryCost, &order.Payment.GoodsTotal,
		&order.Payment.CustomFee)
	if err != nil {
		return models.Order{}, fmt.Errorf("%s: get payment: %w", op, err)
	}

	// 4. Items
	rows, err := s.db.Query(`
		SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
		FROM items WHERE order_uid = $1`, orderID)
	if err != nil {
		return models.Order{}, fmt.Errorf("%s: get items: %w", op, err)
	}
	defer rows.Close()

	order.Items = []models.Item{}
	for rows.Next() {
		var item models.Item
		if err := rows.Scan(&item.ChrtID, &item.TrackNumber, &item.Price, &item.Rid,
			&item.Name, &item.Sale, &item.Size, &item.TotalPrice,
			&item.NmID, &item.Brand, &item.Status); err != nil {
			return models.Order{}, fmt.Errorf("%s: get items: %w", op, err)
		}
		order.Items = append(order.Items, item)
	}

	return order, nil
}

// Close closes the underlying database connection.
// Should be called on application shutdown.
func (s *Storage) Close() error {
	return s.db.Close()
}
