package session

import (
	"context"
	"encoding/json"
	"time"

	myredis "github.com/olaola-chat/rbp-library/redis"

	"github.com/go-redis/redis/v8"
	"github.com/gogf/gf/container/gmap"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/gsession"
	"github.com/gogf/gf/os/gtimer"
)

const RedisRPCCache = "rpc_cache"

// StorageRedis implements the Session Storage interface with redis.
type StorageRedisV8 struct {
	redis         *redis.Client   // Redis client for session storage.
	prefix        string          // Redis key prefix for session id.
	updatingIdMap *gmap.StrIntMap // Updating TTL set for session id.
}

var (
	// DefaultStorageRedisLoopInterval is the interval updating TTL for session ids
	// in last duration.
	DefaultStorageRedisLoopInterval = 10 * time.Second
)

// NewStorageRedis creates and returns a redis storage object for session.
func NewStorageRedisV8() *StorageRedisV8 {
	s := &StorageRedisV8{
		redis:         myredis.RedisClient(RedisRPCCache),
		prefix:        "Session_",
		updatingIdMap: gmap.NewStrIntMap(true),
	}
	ctx := context.Background()
	gtimer.AddSingleton(DefaultStorageRedisLoopInterval, func() {
		var (
			id         string
			err        error
			ttlSeconds int
			num        int
		)
		now := time.Now().UnixNano()
		for {
			if id, ttlSeconds = s.updatingIdMap.Pop(); id == "" {
				break
			} else {
				num++
				if err = s.doUpdateTTL(ctx, id, ttlSeconds); err != nil {
					g.Log().Println(err)
				}
			}
		}
		if num > 0 {
			g.Log().Printf("StorageRedis.timer num=%d used=%f\n", num, float64(time.Now().UnixNano()-now)/1e6)
		}
	})
	return s
}

// New creates a session id.
// This function can be used for custom session creation.
func (s *StorageRedisV8) New(ttl time.Duration) (id string) {
	return ""
}

// Get retrieves session value with given key.
// It returns nil if the key does not exist in the session.
func (s *StorageRedisV8) Get(ctx context.Context, id string, key string) interface{} {
	return nil
}

// GetMap retrieves all key-value pairs as map from storage.
func (s *StorageRedisV8) GetMap(ctx context.Context, id string) map[string]interface{} {
	return nil
}

// GetSize retrieves the size of key-value pairs from storage.
func (s *StorageRedisV8) GetSize(ctx context.Context, id string) int {
	return -1
}

// Set sets key-value session pair to the storage.
// The parameter <ttl> specifies the TTL for the session id (not for the key-value pair).
func (s *StorageRedisV8) Set(ctx context.Context, id string, key string, value interface{}, ttl time.Duration) error {
	return gsession.ErrorDisabled
}

// SetMap batch sets key-value session pairs with map to the storage.
// The parameter <ttl> specifies the TTL for the session id(not for the key-value pair).
func (s *StorageRedisV8) SetMap(ctx context.Context, id string, data map[string]interface{}, ttl time.Duration) error {
	return gsession.ErrorDisabled
}

// Remove deletes key with its value from storage.
func (s *StorageRedisV8) Remove(ctx context.Context, id string, key string) error {
	return gsession.ErrorDisabled
}

// RemoveAll deletes all key-value pairs from storage.
func (s *StorageRedisV8) RemoveAll(ctx context.Context, id string) error {
	return gsession.ErrorDisabled
}

// GetSession returns the session data as *gmap.StrAnyMap for given session id from storage.
//
// The parameter <ttl> specifies the TTL for this session, and it returns nil if the TTL is exceeded.
// The parameter <data> is the current old session data stored in memory,
// and for some storage it might be nil if memory storage is disabled.
//
// This function is called ever when session starts.
func (s *StorageRedisV8) GetSession(ctx context.Context, id string, ttl time.Duration, data *gmap.StrAnyMap) (*gmap.StrAnyMap, error) {
	content, err := s.redis.Get(ctx, s.key(id)).Bytes()
	if err != nil {
		return nil, err
	}
	if len(content) == 0 {
		return nil, nil
	}
	var m map[string]interface{}
	if err = json.Unmarshal(content, &m); err != nil {
		return nil, err
	}
	if m == nil {
		return nil, nil
	}
	if data == nil {
		return gmap.NewStrAnyMapFrom(m, true), nil
	} else {
		data.Replace(m)
	}
	return data, nil
}

// SetSession updates the data map for specified session id.
// This function is called ever after session, which is changed dirty, is closed.
// This copy all session data map from memory to storage.
func (s *StorageRedisV8) SetSession(ctx context.Context, id string, data *gmap.StrAnyMap, ttl time.Duration) error {
	content, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = s.redis.SetEX(ctx, s.key(id), content, ttl).Result()
	return err
}

// UpdateTTL updates the TTL for specified session id.
// This function is called ever after session, which is not dirty, is closed.
// It just adds the session id to the async handling queue.
func (s *StorageRedisV8) UpdateTTL(ctx context.Context, id string, ttl time.Duration) error {
	if ttl >= DefaultStorageRedisLoopInterval {
		s.updatingIdMap.Set(id, int(ttl.Seconds()))
	}
	return nil
}

// doUpdateTTL updates the TTL for session id.
func (s *StorageRedisV8) doUpdateTTL(ctx context.Context, id string, ttlSeconds int) error {
	_, err := s.redis.Expire(ctx, s.key(id), time.Duration(ttlSeconds)).Result()
	return err
}

func (s *StorageRedisV8) key(id string) string {
	return s.prefix + id
}
