package redis

import (
	"github.com/go-redis/redis"
	"log"
	"strconv"
	"sync"
	"time"
)

type Limiter struct {
	mu    sync.Mutex
	key   string
	cli   *redis.Client
	limit float64
	burst int
}

type Options struct {
	Limit  float64
	Burst  int
	Client *redis.Client
	Key    string
}

func NewLimiter(opt *Options) *Limiter {
	return &Limiter{
		limit: opt.Limit,
		burst: opt.Burst,
		cli:   opt.Client,
		key:   opt.Key,
	}
}

func (l *Limiter) Burst() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.burst
}

func (l *Limiter) Limit() float64 {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.limit
}

func (l *Limiter) SetBurst(burst int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.burst = burst
}

func (l *Limiter) SetLimit(limit float64) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.limit = limit
}

func (l *Limiter) Allow() bool {
	return l.AllowN(1)
}

func (l *Limiter) AllowN(n int) bool {
	waitMicrosecond, err := l.ReserveN(n)
	if err != nil {
		return false
	}

	return waitMicrosecond == 0
}

func (l *Limiter) WaitN(n int) bool {
	var waitMicrosecond = time.Duration(1)
	var err error

	for waitMicrosecond > 0 {
		waitMicrosecond, err = l.ReserveN(n)
		if err != nil {
			return false
		}
		time.Sleep(waitMicrosecond)
	}

	return true
}

func (l *Limiter) ReserveN(n int) (time.Duration, error) {
	l.mu.Lock()
	burst := l.burst
	limit := l.limit
	l.mu.Unlock()

	if n > burst {
		panic("n不能大于burst")
	}

	waitMicroSeconds, err := l.cli.Eval(
		script,
		[]string{
			l.key,
			strconv.Itoa(burst),
			strconv.FormatFloat(limit, 'f', 8, 64)},
		n,
	).Float64()

	if err != nil {
		log.Println(err)
	}

	return time.Microsecond * time.Duration(waitMicroSeconds), err
}
