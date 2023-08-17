package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gogf/gf/errors/gerror"
	"github.com/gogf/gf/frame/g"
)

type Mutex interface {
	Lock(ctx context.Context) error
	LockWithTtl(ctx context.Context, ttl time.Duration) error
	UnLock(ctx context.Context) (err error)
}

type mutex struct {
	redis *redis.Client
	key   string
	start time.Time

	maxWaitTime time.Duration
	interval    time.Duration
}

func NewMutex(redisName string, key string) *mutex {
	m := &mutex{
		key:         key,
		redis:       RedisClient(redisName),
		maxWaitTime: time.Second,
		interval:    time.Millisecond,
	}
	return m
}

func (m *mutex) Lock(ctx context.Context) error {
	return m.LockWithTtl(ctx, time.Second*3)
}

func (m *mutex) LockWithTtl(ctx context.Context, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = time.Second * 3
	}
	// TODO: 解决饥饿线程
	waitTime := time.Now()
	for {
		cmd := m.redis.SetNX(ctx, m.key, 1, ttl)
		result, err := cmd.Result()
		if err != nil {
			g.Log().Errorf("Lock Key: %+v, Error: %+v", m.key, err)
			return err
		}
		if result {
			m.start = time.Now()
			g.Log().Debugf("Locked Key : %+v", m.key)
			break
		}
		time.Sleep(m.interval)
		if time.Since(waitTime) > m.maxWaitTime {
			return gerror.New(fmt.Sprintf("Lock Wait Time Too Long %+v", time.Since(waitTime)))
		}
	}
	return nil
}

// 试图获取锁，没有也不阻塞
func (m *mutex) TryLockWithTtl(ctx context.Context, ttl time.Duration) (isSuccess bool, err error) {
	if ttl <= 0 {
		ttl = time.Second * 3
	}
	cmd := m.redis.SetNX(ctx, m.key, 1, ttl)
	result, err := cmd.Result()
	if err != nil {
		g.Log().Errorf("Lock Key: %+v, Error: %+v", m.key, err)
		return false, err
	}
	return result, nil
}

func (m *mutex) UnLock(ctx context.Context) (err error) {
	if m.start.IsZero() {
		g.Log().Error("Unlock is Not locked")
		return nil
	}
	defer func() {
		if err == nil {
			m.start = time.Time{}
		}
	}()
	if time.Since(m.start) > time.Second {
		g.Log().Errorf("Lock Key: %+v, Time : %+v", m.key, time.Since(m.start))
	} else {
		g.Log().Debugf("Lock Key: %+v, Time : %+v", m.key, time.Since(m.start))
	}
	del := m.redis.Del(ctx, m.key)
	result, err := del.Result()
	if err != nil {
		g.Log().Errorf("Unlock Key: %+v, Error: %+v", m.key, err)
		return err
	}
	if result <= 0 {
		g.Log().Errorf("Unlock Key NotExists %+v", m.key)
	}
	return nil
}
