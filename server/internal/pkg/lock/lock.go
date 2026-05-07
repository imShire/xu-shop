// Package lock 提供基于 Redis 的分布式锁。
package lock

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// ErrNotLocked 表示锁不存在（解锁时用）。
var ErrNotLocked = errors.New("lock: key not locked")

// Lock 分布式锁接口。
type Lock interface {
	TryLock(ctx context.Context, key string, ttl time.Duration) (bool, error)
	Unlock(ctx context.Context, key string) error
}

// RedisLock 基于 Redis SET NX 实现的分布式锁。
type RedisLock struct {
	client *redis.Client
}

// New 创建 RedisLock。
func New(client *redis.Client) Lock {
	return &RedisLock{client: client}
}

// TryLock 尝试获取锁，成功返回 true。
func (l *RedisLock) TryLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return l.client.SetNX(ctx, "lock:"+key, 1, ttl).Result()
}

// Unlock 释放锁。
func (l *RedisLock) Unlock(ctx context.Context, key string) error {
	n, err := l.client.Del(ctx, "lock:"+key).Result()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotLocked
	}
	return nil
}
