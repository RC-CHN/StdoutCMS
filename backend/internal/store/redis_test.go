package store

import (
	"context"
	"fmt"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestChatAppendTurn(t *testing.T) {
	addr := os.Getenv("TEST_REDIS_ADDR")
	if addr == "" {
		t.Skip("set TEST_REDIS_ADDR to run Redis integration tests")
	}
	rd, err := NewRedis(addr)
	if err != nil {
		t.Fatal(err)
	}
	defer rd.Close()
	ctx := t.Context()
	sid := uuid.NewString()
	key := "chat:session:" + sid
	defer rd.client.Del(context.Background(), key, key+":ctx")
	var want []string
	for i := 0; i <= chatHistoryLimit; i++ {
		user, assistant := fmt.Sprintf("question %d", i), fmt.Sprintf("reply %d", i)
		want = append(want, user, assistant)
		if err := rd.ChatAppendTurn(ctx, sid, user, assistant, "my-article", time.Hour); err != nil {
			t.Fatal(err)
		}
	}
	got, err := rd.ChatGetHistory(ctx, sid)
	if err != nil || !slices.Equal(got, want[2:]) {
		t.Fatalf("history = %v, error = %v", got, err)
	}
	last, err := rd.ChatGetLastCtx(ctx, sid)
	if err != nil || last != "my-article" {
		t.Fatalf("last article = %q, error = %v", last, err)
	}
}
