package configuration_manager

import (
	"fmt"
	"go-gw-test/pkg/configuration_manager/types"

	"github.com/redis/go-redis/v9"
)

func newRedisClient(redisConfig *types.RedisConfig) *redis.Client {
	if redisConfig.Host == "" || redisConfig.Port == 0 {
		return nil
	}

	return redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port),
		Password: redisConfig.Password,
		DB:       redisConfig.DB,
	})
}
