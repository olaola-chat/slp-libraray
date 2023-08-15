package redis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/extra/rediscmd"
	"github.com/gogf/gf/errors/gerror"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/frame/gins"

	"github.com/opentracing/opentracing-go/log"

	"github.com/go-redis/redis/v8"
	"github.com/opentracing/opentracing-go"

	"github.com/olaola-chat/rbp-library/tracer/wrap"
)

type redisConfig struct {
	Host     string
	Port     int
	Password string
	Db       int
}

// RedisClient 根据name实例化redis对象
func RedisClient(name string) *redis.Client {
	instanceKey := fmt.Sprintf("self-go-redis.%s", name)
	result := gins.GetOrSetFuncLock(instanceKey, func() interface{} {
		config := redisConfig{}
		err := g.Cfg().GetStruct(fmt.Sprintf("go-redis.%s", name), &config)
		if err != nil {
			panic(gerror.Wrap(err, "redis config error"))
		}
		addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
		options := redis.Options{
			Addr:               addr,
			Dialer:             nil,
			OnConnect:          nil,
			Password:           config.Password,
			DB:                 config.Db,
			MaxRetries:         1,
			MinRetryBackoff:    0,
			MaxRetryBackoff:    0,
			DialTimeout:        0,
			ReadTimeout:        0,
			WriteTimeout:       0,
			PoolSize:           0,
			MinIdleConns:       0,
			MaxConnAge:         0,
			PoolTimeout:        0,
			IdleTimeout:        0,
			IdleCheckFrequency: 0,
			TLSConfig:          nil,
		}
		// 新建一个client
		client := redis.NewClient(&options)
		client.AddHook(&tracingHook{})
		return client
	})
	if client, ok := result.(*redis.Client); ok {
		return client
	}
	panic(gerror.New("get redis client error"))
}

type tracingHook struct{}
type tracingRedisStruct struct{}

var (
	tracingRedisKey = tracingRedisStruct{}
)

func (tracingHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	span, spCtx := wrap.StartOpentracingSpan(ctx, fmt.Sprintf("redis-%s", cmd.FullName()))
	if span != nil {
		span.LogFields(
			log.String("db.system", "redis"),
			log.String("db.statement", rediscmd.CmdString(cmd)),
		)
		return context.WithValue(spCtx, tracingRedisKey, true), nil
	}
	return ctx, nil
}

func (hook tracingHook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	span := hook.getStpanFromContext(ctx)
	if span != nil {
		span.Finish()
	}
	return nil
}

func (tracingHook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	summary, cmdsString := rediscmd.CmdsString(cmds)
	span, spCtx := wrap.StartOpentracingSpan(ctx, "redis-pipeline "+summary)
	if span != nil {
		span.LogFields(
			log.String("db.system", "redis"),
			log.Int("db.redis.num_cmd", len(cmds)),
			log.String("db.statement", cmdsString),
		)
		return context.WithValue(spCtx, tracingRedisKey, true), nil
	}

	return ctx, nil
}

func (hook tracingHook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	span := hook.getStpanFromContext(ctx)
	if span != nil {
		span.Finish()
	}
	return nil
}

func (tracingHook) getStpanFromContext(ctx context.Context) opentracing.Span {
	value := ctx.Value(tracingRedisKey)
	if value != nil {
		yes, ok := value.(bool)
		if ok && yes {
			return opentracing.SpanFromContext(ctx)
		}
	}
	return nil
}
