package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

func New(ctx context.Context, addr, password string) (goredis.UniversalClient, error) {
	if addr == "" {
		return nil, fmt.Errorf("redis addr is empty")
	}
	cli := goredis.NewClient(&goredis.Options{
		Addr:         addr,
		Password:     password,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})
	if err := cli.Ping(ctx).Err(); err != nil {
		_ = cli.Close()
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	return cli, nil
}
