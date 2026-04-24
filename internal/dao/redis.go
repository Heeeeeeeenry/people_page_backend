package dao

import (
	"context"
	"fmt"

	"people-page-backend/internal/config"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client
var Ctx = context.Background()

func InitRedis() error {
	cfg := config.AppConfig.Redis
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	RDB = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	_, err := RDB.Ping(Ctx).Result()
	if err != nil {
		return fmt.Errorf("连接Redis失败: %w", err)
	}
	return nil
}

func CloseRedis() {
	if RDB != nil {
		RDB.Close()
	}
}
