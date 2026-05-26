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

// ---- Cache ----

func (r *Redis) GetCache(key string) ([]byte, bool) {
	ctx := context.Background()
	b, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, false
	}
	return b, true
}

func (r *Redis) SetCache(key string, data []byte, ttl time.Duration) {
	ctx := context.Background()
	r.client.Set(ctx, key, data, ttl)
}

func (r *Redis) DeleteCachePattern(pattern string) error {
	ctx := context.Background()
	var cursor uint64
	for {
		keys, next, err := r.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			r.client.Del(ctx, keys...)
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return nil
}

func (r *Redis) InvalidatePostList() {
	r.DeleteCachePattern("posts:list:*")
}

func (r *Redis) InvalidatePost(slug string) {
	ctx := context.Background()
	r.client.Del(ctx, "posts:slug:"+slug)
	r.InvalidatePostList()
}

func (r *Redis) InvalidateProjects() {
	ctx := context.Background()
	r.client.Del(ctx, "projects:list")
}
