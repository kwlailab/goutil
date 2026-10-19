package redisconfig

import (
	"context"

	goredis "github.com/redis/go-redis/v9"
	"github.com/spf13/viper"

	pkgrds "github.com/kwlailab/goutil/pkg/redis"
)

func FromViper(ctx context.Context) (goredis.UniversalClient, error) {
	return pkgrds.New(ctx, viper.GetString("redis.addr"), viper.GetString("redis.password"))
}
