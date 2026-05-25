package store

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
}

func NewRedis(addr string) (*Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}

	return &Redis{client: client}, nil
}

func (r *Redis) Close() error {
	return r.client.Close()
}

func (r *Redis) SetSession(token string, username string, ttl int) error {
	ctx := context.Background()
	return r.client.Set(ctx, "session:"+token, username, time.Duration(ttl)*time.Second).Err()
}

func (r *Redis) ValidateSession(token string) (bool, error) {
	ctx := context.Background()
	_, err := r.client.Get(ctx, "session:"+token).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *Redis) DeleteSession(token string) error {
	ctx := context.Background()
	return r.client.Del(ctx, "session:"+token).Err()
}
