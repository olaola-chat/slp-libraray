package acm

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/olaola-chat/slp-library/consul"

	"github.com/gogf/gf/errors/gerror"
	"github.com/gogf/gf/frame/g"
	"github.com/hashicorp/consul/api"
)

const (
	prefix = "Acm"
)

// acm 内部类型，单例，不允许外部直接创建
type acm struct {
	m        sync.Mutex
	c        *api.Client
	keys     map[string]KeyCallback
	dirs     map[string]DirCallback
	keyIndex map[string]uint64
	dirIndex map[string]uint64
}

func (a *acm) ListenKey(key string, cb KeyCallback) error {
	a.m.Lock()
	defer a.m.Unlock()
	err := a.initFetchKey(key, cb)
	if err != nil {
		return err
	}
	fullKey := fmt.Sprintf("%s/%s", prefix, key)
	a.keys[fullKey] = cb
	return nil
}

func (a *acm) ListenDir(key string, cb DirCallback) error {
	a.m.Lock()
	defer a.m.Unlock()
	err := a.initFetchDir(key, cb)
	if err != nil {
		return err
	}

	fullKey := fmt.Sprintf("%s/%s", prefix, key)
	a.dirs[fullKey] = cb
	return nil
}

func (a *acm) init() {
	a.keys = make(map[string]KeyCallback)
	a.dirs = make(map[string]DirCallback)
	a.keyIndex = make(map[string]uint64)
	a.dirIndex = make(map[string]uint64)

	cfg := &consul.DiscoverConfig{}
	err := g.Cfg().GetStruct("rpc.discover", cfg)
	if err != nil {
		panic(err)
	}
	if cfg.Type == "consul" && len(cfg.Addr) == 0 {
		consulAgentIp := os.Getenv("CONSUL_AGENT_IP")
		if consulAgentIp == "" {
			panic(gerror.Wrap(err, "rpc discover config error"))
		}
		cfg.Addr = []string{fmt.Sprintf("%s:%d", consulAgentIp, 8500)}
	}
	client, err := api.NewClient(&api.Config{
		Address: cfg.Addr[0],
	})
	if err != nil {
		panic(err)
	}
	a.c = client
}

func (a *acm) run() {
	kvClient := a.c.KV()
	var waitIndex uint64 = 0
	for {
		pairs, meta, err := kvClient.List(fmt.Sprintf("%s/", prefix), &api.QueryOptions{
			WaitIndex: waitIndex,
			WaitTime:  time.Second * 60,
		})
		if err != nil || pairs == nil || len(pairs) == 0 {
			g.Log().Error("acm listen error", err)
			time.Sleep(time.Second)
			continue
		}
		a.m.Lock()
		tmpDirIndex := make(map[string]uint64)
		tmpDirData := make(map[string]map[string]string)
		for _, pair := range pairs {
			key := pair.Key
			if callback, ok := a.keys[key]; ok {
				if preIndex, ok := a.keyIndex[key]; !ok || pair.ModifyIndex > preIndex {
					g.Log().Info("acm key changed", key, pair.ModifyIndex, preIndex)
					err := callback(string(pair.Value))
					if err != nil {
						g.Log().Error("Acm auto listener callback error", err)
					} else {
						a.keyIndex[key] = pair.ModifyIndex
					}
				}
			}

			for k := range a.dirs {
				if strings.HasPrefix(key, k) {
					if i, ok := tmpDirIndex[k]; ok {
						if i < pair.ModifyIndex {
							tmpDirIndex[k] = pair.ModifyIndex
						}
					} else {
						tmpDirIndex[k] = pair.ModifyIndex
					}

					if d, ok := tmpDirData[k]; ok {
						d[pair.Key] = string(pair.Value)
					} else {
						tmpDirData[k] = make(map[string]string)
						tmpDirData[k][pair.Key] = string(pair.Value)
					}
				}
			}
		}

		for dir, index := range tmpDirIndex {
			if preIndex, ok := a.dirIndex[dir]; !ok || preIndex < index {
				callback, ok := a.dirs[dir]
				if !ok {
					continue
				}

				err := callback(tmpDirData[dir])
				if err != nil {
					g.Log().Error("Acm auto listener dir callback error", err)
				} else {
					a.dirIndex[dir] = index
				}
			}
		}

		waitIndex = meta.LastIndex
		a.m.Unlock()
	}
}

func (a *acm) initFetchKey(key string, callback KeyCallback) error {
	fullKey := fmt.Sprintf("%s/%s", prefix, key)
	pair, meta, err := a.c.KV().Get(fullKey, &api.QueryOptions{})
	if err != nil {
		g.Log().Error("acm get", key, "error", err)
		return err
	}
	if pair == nil {
		g.Log().Error("acm get", key, "not found")
		return gerror.Newf("key not found %s", fullKey)
	}
	g.Log().Info("acm get ok", key, meta.LastIndex)
	err = callback(string(pair.Value))
	if err != nil {
		g.Log().Error("acm get error ", err)
		return err
	}
	a.keyIndex[fullKey] = meta.LastIndex

	return nil
}

func (a *acm) initFetchDir(dir string, callback DirCallback) error {
	fullKey := fmt.Sprintf("%s/%s", prefix, dir)
	pairs, meta, err := a.c.KV().List(fullKey, &api.QueryOptions{})
	if err != nil {
		g.Log().Error("acm get error", err)
		return err
	}
	dirData := make(map[string]string)
	var lastIndex uint64
	for _, pair := range pairs {
		key := pair.Key
		dirData[key] = string(pair.Value)
		if lastIndex < pair.ModifyIndex {
			lastIndex = pair.ModifyIndex
		}
	}

	g.Log().Info("acm get ok", dir, meta.LastIndex)
	err = callback(dirData)
	if err != nil {
		g.Log().Error("acm get error ", err)
		return err
	}
	a.dirIndex[fullKey] = lastIndex
	return nil
}
