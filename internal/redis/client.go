package redis

import (
	"context"
	"fmt"
	"net"

	"github.com/MoSed3/otp-server/internal/config"
	"github.com/redis/go-redis/v9"
)

type Config struct {
	client *redis.Client
	ctx    context.Context
	cancel context.CancelFunc
}

func New(redisConfig config.RedisConfig) *Config {
	ctx, cancel := context.WithCancel(context.Background())
	c := &Config{
		client: redis.NewClient(&redis.Options{
			Addr:     net.JoinHostPort(redisConfig.Host, fmt.Sprintf("%d", redisConfig.Port)),
			Password: redisConfig.Password,
			DB:       redisConfig.DB,
		}),
		ctx:    ctx,
		cancel: cancel,
	}
	return c
}

func (c *Config) resetCtx() {
	c.ctx, c.cancel = context.WithCancel(context.Background())
}

func (c *Config) Start() error {
	_, err := c.client.Ping(c.ctx).Result()
	return err
}

func (c *Config) Stop() {
	defer c.cancel()
	defer c.resetCtx()

	if c.client != nil {
		err := c.client.Close()
		if err != nil {
			fmt.Printf("Error closing Redis client: %v\n", err)
		} else {
			fmt.Println("Redis client disconnected successfully.")
		}
	}
}
