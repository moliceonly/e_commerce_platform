package cache

import (
	"context"
	"fmt"
	"time"
)

// Cache 缓存端口（阶段 G · Redis）。
// 建议用途（实现时任选组合，不要只做下单读库）：
//   - 商品详情：key product:{id}
//   - 商品列表短缓存：key products:page:{n}:size:{m}
//   - 用户购物车摘要：key cart:{userID}
//   - 登录限流 / token 黑名单：key login:fail:{email} / jwt:deny:{jti}
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, val string, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) error
	Ping(ctx context.Context) error
}

// RedisCache go-redis 封装骨架。
// 依赖：go get github.com/redis/go-redis/v9
type RedisCache struct {
	// Client *redis.Client // TODO(G): 填入客户端
	Addr string
}

// NewRedis 连接 Redis。addr 形如 127.0.0.1:6379 或 redis:6379。
func NewRedis(addr string) (*RedisCache, error) {
	// TODO(G):
	//  rdb := redis.NewClient(&redis.Options{Addr: addr})
	//  if err := rdb.Ping(ctx).Err(); err != nil { return nil, err }
	//  return &RedisCache{Client: rdb, Addr: addr}, nil
	_ = addr
	return nil, fmt.Errorf("TODO(G): cache.NewRedis not implemented")
}

func (c *RedisCache) Get(ctx context.Context, key string) (string, error) {
	// TODO(G): return c.Client.Get(ctx, key).Result()；区分 redis.Nil
	return "", fmt.Errorf("TODO(G): RedisCache.Get")
}

func (c *RedisCache) Set(ctx context.Context, key, val string, ttl time.Duration) error {
	// TODO(G): return c.Client.Set(ctx, key, val, ttl).Err()
	return fmt.Errorf("TODO(G): RedisCache.Set")
}

func (c *RedisCache) Del(ctx context.Context, keys ...string) error {
	// TODO(G): return c.Client.Del(ctx, keys...).Err()
	return fmt.Errorf("TODO(G): RedisCache.Del")
}

func (c *RedisCache) Ping(ctx context.Context) error {
	// TODO(G): return c.Client.Ping(ctx).Err()
	return fmt.Errorf("TODO(G): RedisCache.Ping")
}

// 常用 key 约定（可按需改）。
func ProductKey(id uint) string          { return fmt.Sprintf("product:%d", id) }
func ProductsPageKey(page, size int) string {
	return fmt.Sprintf("products:page:%d:size:%d", page, size)
}
func CartKey(userID uint) string { return fmt.Sprintf("cart:%d", userID) }
