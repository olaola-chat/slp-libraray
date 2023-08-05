package library

import (
	"context"
	"fmt"

	"github.com/olaola-chat/rbp-library/tracer/wrap"

	"github.com/go-redis/redis/extra/rediscmd"
	"github.com/go-redis/redis/v8"
	"github.com/gogf/gf/errors/gerror"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/frame/gins"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

const (
	//RedisDefault 默认的redis服务器
	//其实他并不默认，使用时，调用者必须清楚的知晓存储数据的规模以及请求量
	RedisDefault = "default"
	//RedisCache 一般缓存redis
	RedisCache = "cache"
	//RedisCacheOld 一般缓存redis
	RedisCacheOld = "cache_old"
	//RedisRoom 房间缓存Redis
	RedisRoom = "room"
	//RedisUser 用户关系缓存
	RedisUser = "user"
	//RedisMate
	RedisMate = "mate"
	//RedisRPCCache 用户信息服务，认证服务
	RedisRPCCache = "rpc_cache"
	//RedisMatch 匹配服务
	RedisMatch = "match"
	//RedisSearch 自建redis，用于房间检索，搜索入口，附近...
	RedisSearch = "search"
	//曝光和点击记录
	RedisRecord = "record"
	//redisEs队列
	RedisEs = "es"
)

type redisConfig struct {
	Host     string
	Port     int
	Password string
}

type RedisInstance struct {
	RedisDefault  *redis.Client
	RedisCache    *redis.Client
	RedisCacheOld *redis.Client
	RedisRoom     *redis.Client
	RedisUser     *redis.Client
	RedisMate     *redis.Client
	RedisRPCCache *redis.Client
	RedisMatch    *redis.Client
	RedisSearch   *redis.Client
	RedisRecord   *redis.Client
	RedisEs       *redis.Client
}

var instance *RedisInstance

// 初始化redis实例
func init() {
	g.Log().Info("init redis instance")

	instance = &RedisInstance{
		RedisDefault:  RedisClientInit(RedisDefault),
		RedisCache:    RedisClientInit(RedisCache),
		RedisCacheOld: RedisClientInit(RedisCacheOld),
		RedisRoom:     RedisClientInit(RedisRoom),
		RedisUser:     RedisClientInit(RedisUser),
		RedisMate:     RedisClientInit(RedisMate),
		RedisRPCCache: RedisClientInit(RedisRPCCache),
		RedisMatch:    RedisClientInit(RedisMatch),
		RedisSearch:   RedisClientInit(RedisSearch),
		RedisRecord:   RedisClientInit(RedisRecord),
		RedisEs:       RedisClientInit(RedisEs),
	}
}

// RedisClient 根据name获取redis对象
func RedisClient(name string) *redis.Client {
	switch name {
	case RedisDefault:
		return instance.RedisDefault
	case RedisCache:
		return instance.RedisCache
	case RedisCacheOld:
		return instance.RedisCacheOld
	case RedisRoom:
		return instance.RedisRoom
	case RedisUser:
		return instance.RedisUser
	case RedisMate:
		return instance.RedisMate
	case RedisRPCCache:
		return instance.RedisRPCCache
	case RedisMatch:
		return instance.RedisMatch
	case RedisSearch:
		return instance.RedisSearch
	case RedisRecord:
		return instance.RedisRecord
	case RedisEs:
		return instance.RedisEs
	default:
		panic(gerror.New("get redis client error"))
	}
}

// RedisClientInit 根据name实例化redis对象
func RedisClientInit(name string) *redis.Client {
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
			DB:                 0,
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
