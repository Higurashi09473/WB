package redis

import (
	"WB/internal/config"
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	Client *redis.Client
}

func New(cfg *config.Config) (*Redis, error) {
	const op = "storage.redis.New"

	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password:     cfg.Redis.Password,
		DB:           cfg.DB,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
        client.Close() 
        return nil, fmt.Errorf("%s: ping failed: %w", op, err)
    }
	
	return &Redis{Client: client}, nil
}

func (r *Redis) Close() error{
	return r.Client.Close()
}


func (r *Redis) GetOrder(ctx context.Context, orderUID string) ([]byte, error) {
	key := orderUID

	data, err := r.Client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	return data, nil
}

func (r *Redis) SetOrder(ctx context.Context, orderUID string, data []byte, ttl time.Duration) error {
	key := orderUID

	err := r.Client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *Redis) DeleteOrder(ctx context.Context, orderUID string) error {
	key := orderUID

	err := r.Client.Del(ctx, key).Err()
	if err != nil {
		return err
	}

	return nil
}