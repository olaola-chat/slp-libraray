package acm

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/olaola-chat/rbp-library/consul"

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

func (a *acm) ListenKey(key string, cb KeyCallback) {
	a.m.Lock()
	defer a.m.Unlock()

	a.keys[key] = cb
}

func (a *acm) ListenDir(key string, cb DirCallback) {
	a.m.Lock()
	defer a.m.Unlock()

	a.dirs[key] = cb
}

func (a *acm) init() {
	cfg := &consul.DiscoverConfig{}
	err := g.Cfg().GetStruct("rpc.discover", cfg)
	if err != nil {
		panic(err)
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
