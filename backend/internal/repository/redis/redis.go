package redis

import (
	"WB/internal/config"
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	Client *redis.Client
}

func New(cfg *config.Config) (*Redis, error) {
	const op = "storage.redis.New"

	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password:     cfg.Password,
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


//нету
func (r *Redis) GetOrder(ctx context.Context, orderUID string) ([]byte, error){
	return nil, nil
}