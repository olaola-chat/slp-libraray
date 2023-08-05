package tool

import (
	"context"
	"encoding/json"
	"reflect"
	"time"

	redis2 "github.com/go-redis/redis/v8"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/util/gconv"
	"google.golang.org/protobuf/proto"
)

var Redis = &redis{}

type redis struct{}

func (r *redis) FetchWithProtobufExpire(ctx context.Context, redisCli *redis2.Client, redisKey string,
	fn func() (proto.Message, time.Duration, error), typ reflect.Type) (proto.Message, error) {

	val, err := redisCli.Get(ctx, redisKey).Bytes()
	if err != nil {
		if err == redis2.Nil {
			tmp, redisExpire, err := fn()
			if err != nil {
				return nil, err
			}
			var ret = reflect.New(typ)
			if bb, err := proto.Marshal(tmp); err != nil {
				g.Log().Error("proto.Marshal failed", err)
				return nil, err
			} else {
				if err := redisCli.Set(ctx, redisKey, bb, redisExpire).Err(); err != nil {
					g.Log().Errorf("redis.Set(%v) failed: %v", redisKey, err)
				}

				if err = proto.Unmarshal(bb, ret.Interface().(proto.Message)); err != nil {
					return nil, err
				}
			}
			return ret.Interface().(proto.Message), nil
		} else {
			g.Log().Errorf("redisCli.Get(%v) failed: %v", redisKey, err)
			return nil, err
		}
	} else {
		var tmp = reflect.New(typ)
		err := proto.Unmarshal(val, tmp.Interface().(proto.Message))
		if err != nil {
			g.Log().Error("fail to proto.Unmarshal", err)
			return nil, err
		}
		return tmp.Interface().(proto.Message), nil
	}
}

func (r *redis) FetchWithProtobuf(ctx context.Context, redisCli *redis2.Client, redisKey string, redisExpire time.Duration,
	fn func() (proto.Message, error), typ reflect.Type) (proto.Message, error) {

	val, err := redisCli.Get(ctx, redisKey).Bytes()
	if err != nil {
		if err == redis2.Nil {
			tmp, err := fn()
			if err != nil {
				return nil, err
			}
			var ret = reflect.New(typ)
			if bb, err := proto.Marshal(tmp); err != nil {
				g.Log().Error("proto.Marshal failed", err)
				return nil, err
			} else {
				if err := redisCli.Set(ctx, redisKey, bb, redisExpire).Err(); err != nil {
					g.Log().Errorf("redis.Set(%v) failed: %v", redisKey, err)
				}

				if err = proto.Unmarshal(bb, ret.Interface().(proto.Message)); err != nil {
					return nil, err
				}
			}
			return ret.Interface().(proto.Message), nil
		} else {
			g.Log().Errorf("redisCli.Get(%v) failed: %v", redisKey, err)
			return nil, err
		}
	} else {
		var tmp = reflect.New(typ)
		err := proto.Unmarshal(val, tmp.Interface().(proto.Message))
		if err != nil {
			g.Log().Error("fail to proto.Unmarshal", err)
			return nil, err
		}
		return tmp.Interface().(proto.Message), nil
	}
}

func (r *redis) FetchWithString(ctx context.Context, redisCli *redis2.Client, redisKey string, redisExpire time.Duration,
	fn func() (string, error)) (string, error) {

	val, err := redisCli.Get(ctx, redisKey).Result()
	if err != nil {
		if err == redis2.Nil {
			tmp, err := fn()
			if err != nil {
				return "", err
			}

			if err := redisCli.Set(ctx, redisKey, tmp, redisExpire).Err(); err != nil {
				g.Log().Errorf("redis.Set(%v) failed: %v", redisKey, err)
			}
			return tmp, nil

		} else {
			g.Log().Errorf("redisCli.Get(%v) failed: %v", redisKey, err)
			return "", err
		}
	} else {
		return val, nil
	}
}

// FetchWithUint32Expire not test
//func (r *redis) FetchWithUint32Expire(ctx context.Context, redisCli *redis2.Client, redisKey string,
//	fn func() (uint32, time.Duration, error)) (uint32, error) {
//
//	val, err := redisCli.Get(ctx, redisKey).Result()
//	if err != nil {
//		if err == redis2.Nil {
//			tmp, expire, err := fn()
//			if err != nil {
//				return 0, err
//			}
//
//			if err := redisCli.Set(ctx, redisKey, tmp, expire).Err(); err != nil {
//				g.Log().Errorf("redis.Set(%v) failed: %v", redisKey, err)
//			}
//			return tmp, nil
//
//		} else {
//			g.Log().Errorf("redisCli.Get(%v) failed: %v", redisKey, err)
//			return 0, err
//		}
//	} else {
//		return gconv.Uint32(val), nil
//	}
//}

// FetchWithUint32 test
func (r *redis) FetchWithUint32(ctx context.Context, redisCli *redis2.Client, redisKey string, redisExpire time.Duration,
	fn func() (uint32, error)) (uint32, error) {

	val, err := redisCli.Get(ctx, redisKey).Result()
	if err != nil {
		if err == redis2.Nil {
			tmp, err := fn()
			if err != nil {
				return 0, err
			}

			if err := redisCli.Set(ctx, redisKey, tmp, redisExpire).Err(); err != nil {
				g.Log().Errorf("redis.Set(%v) failed: %v", redisKey, err)
			}
			return tmp, nil

		} else {
			g.Log().Errorf("redisCli.Get(%v) failed: %v", redisKey, err)
			return 0, err
		}
	} else {
		return gconv.Uint32(val), nil
	}
}

// FetchWithJson fn返回的interface{} 需要是指针类型，FetchWithJson返回的 interface{} 也是指针类型,
func (r *redis) FetchWithJson(
	ctx context.Context,
	redisCli *redis2.Client,
	redisKey string,
	redisExpire time.Duration,
	fn func() (interface{}, error),
	typ reflect.Type) (interface{}, error) {

	val, err := redisCli.Get(ctx, redisKey).Bytes()

	if err != nil {

		if err == redis2.Nil {

			tmp, err := fn()

			if err != nil {
				return nil, err
			}

			if bb, err := json.Marshal(tmp); err != nil {

				g.Log().Error("json.Marshal failed", err)
				return nil, err

			} else {

				var ret = reflect.New(typ)

				if err := redisCli.Set(ctx, redisKey, bb, redisExpire).Err(); err != nil {
					g.Log().Errorf("redis.Set(%v) failed: %v", redisKey, err)
				}

				if err := json.Unmarshal(bb, ret.Interface()); err != nil {
					g.Log().Error("json.Unmarshal failed", err)
					return nil, err
				}

				return ret.Interface(), nil
			}

		} else {

			g.Log().Errorf("redisCli.Get(%v) failed: %v", redisKey, err)
			return nil, err

		}

	} else {

		var ret = reflect.New(typ)

		err := json.Unmarshal(val, ret.Interface())
		if err != nil {
			g.Log().Error("fail to json.Unmarshal", err)
			return nil, err
		}

		return ret.Interface(), nil

	}
}
