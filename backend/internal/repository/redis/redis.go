package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	Client *redis.Client
}

func New(host, port, password string, DB int) (*Redis, error) {
	const op = "storage.redis.New"

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       DB,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("%s: ping failed: %w", op, err)
	}

	return &Redis{Client: client}, nil
}

func (r *Redis) Close() error {
	return r.Client.Close()
}

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

func (r *Redis) SetOrder(ctx context.Context, orderUID string, data []byte, ttl time.Duration) error {
	const op = "storage.redis.SetOrder"

	key := orderUID

	err := r.Client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("%s: set failed: %w", op, err)
	}

	return nil
}

func (r *Redis) DeleteOrder(ctx context.Context, orderUID string) error {
	const op = "storage.redis.DeleteOrder"

	key := orderUID

	err := r.Client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("%s: del failed: %w", op, err)
	}

	return nil
}
