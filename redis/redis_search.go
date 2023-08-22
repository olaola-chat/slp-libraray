package redis

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/gogf/gf/util/gconv"
	room2 "github.com/olaola-chat/rbp-proto/gen_pb/es/room"

	"github.com/olaola-chat/rbp-library/tool"
)

var searchRedisClient *redis.Client

const (
	fieldTokenization   = ",.<>{}[]\"':;!@#$%^&*()-+=~|"
	RedisInstanceSearch = "search"
)

var RedisSearch = &redisSearch{}

type redisSearch struct {
}

func init() {
	searchRedisClient = RedisClient(RedisInstanceSearch)
}

// BuildKeywordBak 对关键词进行
func BuildKeywordBak(keyword string) string {
	keyword = EscapeTextFileString(strings.TrimSpace(keyword))
	cleanName := tool.Str.CleanString(keyword)
	if len(cleanName) > 0 {
		return fmt.Sprintf("(%s) | %s", strings.Join(strings.Split(cleanName, ""), " "), keyword)
	} else {
		return keyword
	}
}

func EscapeTextFileString(value string) string {
	for _, char := range fieldTokenization {
		value = strings.Replace(value, string(char), "", -1)
	}
	return value
}

// QueryNoContentBak 不返回具体内容，只返回主键(出去前缀)
func (*redisSearch) QueryNoContentBak(ctx context.Context, index string, prefixLen int, params []interface{}, reply *room2.RepEsRoomSearchDefault) error {
	args := []interface{}{"FT.SEARCH", index}
	args = append(args, params...)
	args = append(args, "NOCONTENT")
	result, err := searchRedisClient.Do(ctx, args...).Result()
	if err != nil {
		return err
	}
	res := result.([]interface{})
	reply.Total = gconv.Uint32(res[0])
	for i := 1; i < len(res); i++ {
		key := gconv.String(res[i])
		pk := gconv.Uint32(key[prefixLen:])
		if pk > 0 {
			reply.Data = append(reply.Data, pk)
		}
	}
	return nil
}

// QueryWithContentBak 返回所有数据内容，需要自己对返回内容进行处理
func (*redisSearch) QueryWithContentBak(ctx context.Context, index string, params []interface{}) (uint32, []map[string]string, error) {
	args := []interface{}{"FT.SEARCH", index}
	args = append(args, params...)

	result, err := searchRedisClient.Do(ctx, args...).Result()
	if err != nil {
		return 0, nil, err
	}
	res := result.([]interface{})
	response := make([]map[string]string, 0)
	for i := 1; i < len(res); i += 2 {
		item, ok := res[i+1].([]interface{})
		if ok {
			val := make(map[string]string)
			for j := 0; j < len(item); j += 2 {
				val[item[j].(string)] = item[j+1].(string)
			}
			response = append(response, val)
		}
	}
	return gconv.Uint32(res[0]), response, nil
}

func GetCleanName(currentName string) string {
	cleanName := tool.Str.CleanString(currentName)
	var name string
	if len(cleanName) > 0 {
		name = fmt.Sprintf("%s %s", strings.Join(strings.Split(cleanName, ""), " "), currentName)
	} else {
		name = currentName
	}
	return name
}
