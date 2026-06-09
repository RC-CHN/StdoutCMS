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

// ---- Session ----

func (r *Redis) SetSession(ctx context.Context, token string, username string, ttl int) error {
	return r.client.Set(ctx, "session:"+token, username, time.Duration(ttl)*time.Second).Err()
}

func (r *Redis) ValidateSession(ctx context.Context, token string) (bool, error) {
	_, err := r.client.Get(ctx, "session:"+token).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *Redis) DeleteSession(ctx context.Context, token string) error {
	return r.client.Del(ctx, "session:"+token).Err()
}

// ---- Cache ----

func (r *Redis) GetCache(ctx context.Context, key string) ([]byte, bool) {
	b, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, false
	}
	return b, true
}

func (r *Redis) SetCache(ctx context.Context, key string, data []byte, ttl time.Duration) {
	r.client.Set(ctx, key, data, ttl)
}

func (r *Redis) DeleteCachePattern(ctx context.Context, pattern string) error {
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

func (r *Redis) InvalidatePostList(ctx context.Context) {
	_ = r.DeleteCachePattern(ctx, "posts:list:*")
}

func (r *Redis) InvalidatePost(ctx context.Context, slug string) {
	r.client.Del(ctx, "posts:slug:"+slug)
	r.InvalidatePostList(ctx)
}

func (r *Redis) InvalidateProjects(ctx context.Context) {
	r.client.Del(ctx, "projects:list")
	_ = r.DeleteCachePattern(ctx, "projects:id:*")
}

func (r *Redis) InvalidateAbout(ctx context.Context) {
	r.client.Del(ctx, "about:page")
}

// ---- Chat ----

const chatHistoryLimit = 20 // max rounds per session

// ChatGetHistory returns chat messages from oldest to newest.
// Each element is a JSON blob: {"role":"user","content":"..."}
func (r *Redis) ChatGetHistory(ctx context.Context, sid string) ([]string, error) {
	key := "chat:session:" + sid
	// LRANGE returns newest-first, we want oldest-first
	items, err := r.client.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, err
	}
	// reverse to chronological order (oldest first)
	n := len(items)
	result := make([]string, n)
	for i, v := range items {
		result[n-1-i] = v
	}
	return result, nil
}

// ChatPushMessage pushes a JSON-encoded message onto the session list,
// trims to chatHistoryLimit rounds (×2 messages), and refreshes TTL.
func (r *Redis) ChatPushMessage(ctx context.Context, sid string, msg string, ttl time.Duration) error {
	key := "chat:session:" + sid
	pipe := r.client.Pipeline()
	pipe.LPush(ctx, key, msg)
	pipe.LTrim(ctx, key, 0, chatHistoryLimit*2-1) // 2 messages per round
	pipe.Expire(ctx, key, ttl)
	_, err := pipe.Exec(ctx)
	return err
}

// ChatDailyQuota increments today's global counter and returns (current, limit).
// If the counter was just created, sets its TTL to the end of the day.
func (r *Redis) ChatDailyQuota(ctx context.Context, limit int) (int64, bool) {
	key := "chat:quota:daily"
	n, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, false // Redis down? deny
	}
	if n == 1 {
		// first request today, set TTL to midnight
		now := time.Now()
		midnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		r.client.ExpireAt(ctx, key, midnight)
	}
	return n, n <= int64(limit)
}

// ChatGetLastCtx returns the article slug last injected for this session.
func (r *Redis) ChatGetLastCtx(ctx context.Context, sid string) (string, error) {
	return r.client.Get(ctx, "chat:session:"+sid+":ctx").Result()
}

// ChatSetLastCtx records the current article slug for this session with the same TTL as the session.
func (r *Redis) ChatSetLastCtx(ctx context.Context, sid string, slug string, ttl time.Duration) error {
	return r.client.Set(ctx, "chat:session:"+sid+":ctx", slug, ttl).Err()
}
