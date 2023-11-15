package checks

import (
	"context"
	"fmt"
	"strings"

	redis "github.com/go-redis/redis"
)

// Config is the Redis checker configuration settings container.
type RedisCheck struct {
	Ctx              context.Context
	ConnectionString string
	Error            error
}

func NewRedisCheck(ctx context.Context, connectionString string) RedisCheck {
	return RedisCheck{
		Ctx:              ctx,
		ConnectionString: connectionString,
		Error:            nil,
	}
}

func CheckRedisStatus(ctx context.Context, connectionString string) error {
	check := NewRedisCheck(ctx, connectionString)
	return check.CheckStatus()
}

func (check RedisCheck) CheckStatus() error {
	// support all DSN formats (for backward compatibility) - with and w/out schema and path part:
	// - redis://localhost:1234/
	// - rediss://localhost:1234/
	// - localhost:1234
	redisDSN := check.ConnectionString
	if !strings.HasPrefix(redisDSN, "redis://") && !strings.HasPrefix(redisDSN, "rediss://") {
		redisDSN = fmt.Sprintf("redis://%s", redisDSN)
	}
	redisOptions, _ := redis.ParseURL(redisDSN)
	//ctx := check.Ctx

	rdb := redis.NewClient(redisOptions)
	defer rdb.Close()

	pong, err := rdb.Ping().Result()
	if err != nil {
		check.Error = fmt.Errorf("redis ping failed: %w", err)
		return check.Error
	}

	if pong != "PONG" {
		check.Error = fmt.Errorf("unexpected response for redis ping: %q", pong)
		return check.Error
	}

	return nil
}
