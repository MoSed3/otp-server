package redis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
)

var (
	client      *redis.Client
	ctx, cancel = context.WithCancel(context.Background())
)

func Init(addr, password string, db int) error {
	client = redis.NewClient(&redis.Options{
		Addr:     addr,
		DB:       db,
		Password: password,
	})

	_, err := client.Ping(ctx).Result()
	return err
}

func Disconnect() {
	defer cancel()

	if client != nil {
		err := client.Close()
		if err != nil {
			fmt.Printf("Error closing Redis client: %v\n", err)
		} else {
			fmt.Println("Redis client disconnected successfully.")
		}
	}
}
