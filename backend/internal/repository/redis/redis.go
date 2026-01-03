// Package redis provides Redis-based implementation of order storage.
// It handles database connection and order CRUD operations.
package redis

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

// Redis represents Redis repository for orders.
type Redis struct {
	Client *redis.Client
}

// MustLoad initializes Redis storage with database connection.
func MustLoad(log *slog.Logger, host, port, password string, DB int) *Redis {
	const op = "storage.redis.MustLoad"

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       DB,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		client.Close()
		log.Error("%s: apply migrations: %v", op, err)
		os.Exit(1)
	}

	return &Redis{Client: client}
}

// GetOrder retrieves an order by its orderUID from the database.
func (r *Redis) GetOrder(ctx context.Context, orderUID string) ([]byte, error) {
	const op = "storage.redis.GetOrder"

	key := orderUID

	data, err := r.Client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf("%s: get failed: %w", op, err)
	}

	return data, nil
}

// SetOrder adds a new order to the Redis or returns an error.
func (r *Redis) SetOrder(ctx context.Context, orderUID string, data []byte, ttl time.Duration) error {
	const op = "storage.redis.SetOrder"

	key := orderUID

	err := r.Client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("%s: set failed: %w", op, err)
	}

	return nil
}

// DeleteOrder deletes the order from Redis or returns an error.
func (r *Redis) DeleteOrder(ctx context.Context, orderUID string) error {
	const op = "storage.redis.DeleteOrder"

	key := orderUID

	err := r.Client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("%s: del failed: %w", op, err)
	}

	return nil
}

// Close closes the underlying database connection.
// Should be called on application shutdown.
func (r *Redis) Close() error {
	return r.Client.Close()
}
