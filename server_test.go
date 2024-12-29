package main

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// Test Redis Client Running Alongside Go Server
func TestOfficialRedisClient(t *testing.T) {
	listenAddr := ":5001"
	server := NewServer(ServerConfig{
		ListenAddr: listenAddr,
	})
	go func() {
		defer server.Shutdown()
		log.Fatal(server.Start())
	}()

	time.Sleep(time.Second)

	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("localhost%s", listenAddr),
		Password:     "",
		DB:           0,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	})

	entries := map[string]string{
		"foo": "bar",
	}

	for key, val := range entries {
		if err := rdb.Set(context.Background(), key, val, 0).Err(); err != nil {
			t.Fatal(err)
		}

		newVal, err := rdb.Get(context.Background(), key).Result()
		if err != nil {
			t.Fatal(err)
		}

		if newVal != val {
			t.Fatalf("expected %s but got %s", val, newVal)
		}

		if err = rdb.Del(context.Background(), key).Err(); err != nil {
			t.Fatal(err)
		}

		newVal, err = rdb.Get(context.Background(), key).Result()
		if err != nil {
			t.Fatal(err)
		}

		if newVal != "(nil)" {
			t.Fatalf("expected key %s to be deleted, but got: %s", key, newVal)
		}
	}

}
