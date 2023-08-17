//内存管理基于redis的再goframe的社区版本中使用发现goframe的redis跟业务实际使用不一致

package redis

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/gogf/gf/frame/g"
	"time"
)

// 1.如果redis里面有，就直接返回结果
// 2.如果redis缓存里面没有，就通过闭包获取结果，并存入redis,再返回
// 3. 如果redis不存在，就将空字符串存入 防止击穿
func GetOrSetFunc(ctx context.Context, rds *redis.Client, key string, f func() (bData []byte, err error), duration time.Duration) (data []byte, err error) {
	v, err := rds.Get(ctx, key).Result()
	if err != nil && err != redis.Nil {
		return data, err
	}
	if err == redis.Nil {
		data, err = f()
		if err != nil {
			return []byte{}, err
		}
		return data, rds.Set(ctx, key, string(data), duration).Err()
	} else {
		return []byte(v), err
	}
}

func HashGetOrSetFunc(ctx context.Context, rds *redis.Client, key string, field string, f func() (bData []byte, err error), duration time.Duration) (data []byte, err error) {
	g.Log().Debug("msg", "HashGetOrSetFunc", "err", err, "key", key, "field", field)

	v, err := rds.HGet(ctx, key, field).Result()
	if err != nil && err != redis.Nil {
		return data, err
	}
	if err == redis.Nil {
		//key不存在或者field不存在返回的都是redis.Nil
		data, err = f()
		if err != nil {
			g.Log().Error("msg", "HashGetOrSetFunc", "err", err, "key", key, "field", field)
			return []byte{}, err
		}
		rds.HSet(ctx, key, field, string(data))
		rds.Expire(ctx, key, duration)
		return
	} else {
		return []byte(v), err
	}
}

func DelayDel(ctx context.Context, rds *redis.Client, key string) (err error) {
	_, err = rds.Expire(ctx, key, time.Second).Result()
	if err != nil {
		return err
	}
	return
}

func Del(ctx context.Context, rds *redis.Client, key string) (err error) {
	_, err = rds.Del(ctx, key).Result()
	if err != nil {
		return err
	}
	return
}

func HashDel(ctx context.Context, rds *redis.Client, key string, field string) (err error) {
	_, err = rds.HDel(ctx, key, field).Result()
	if err != nil {
		return err
	}
	return
}
