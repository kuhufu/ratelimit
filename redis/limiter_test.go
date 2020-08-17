package redis

import (
	"fmt"
	"github.com/go-redis/redis"
	"testing"
	"time"
)

func TestLimiter_Allow(t *testing.T) {
	limiter := NewLimiter(&Options{
		Limit: 1,
		Burst: 1,
		Client: redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		}),
		Key: "redis_limit_key",
	})

	for {
		fmt.Println(limiter.AllowN(5))
		time.Sleep(time.Second / 4)
	}
}
