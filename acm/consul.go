package acm

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gogf/gf/errors/gerror"
	"github.com/gogf/gf/frame/g"
	consulapi "github.com/hashicorp/consul/api"
	"gopkg.in/yaml.v3"
)

type discoverConfig struct {
	Type string
	Addr []string
	Path string
}

const (
	prefix = "Acm"
)

// Acm 是一个获取配置的单列
var Acm = &consul{
	listens: make(map[string]acmOnChange),
	changed: make(map[string]uint64),
}

var queue map[string]*config = make(map[string]*config)
var ready = false

// ParseContent 回调解析函数
type ParseContent func(string) error

type config struct {
	DataID          string
	Callback        ParseContent
	IgnoreInitError bool
}

// Init 进程启动时自动初始化，不要多次调用
func Init() {
	for _, config := range queue {
		initConfig(config)
	}
	queue = make(map[string]*config)
	ready = true
	go Acm.Auto()
}

func initConfig(config *config) {
	content, err := Acm.Get(config.DataID)
	if !config.IgnoreInitError {
		if err != nil {
			panic(err)
		}
		err = config.Callback(content)
	} else {
		if err == nil {
			err = config.Callback(content)
		}
	}

	if err != nil {
		g.Log().Error("acm init key error", config.DataID, err)
		panic(err)
	}

	Acm.Listen(config.DataID, func(dataID, content string) error {
		g.Log().Println("Acm Listen", dataID)
		if len(content) > 0 {
			return config.Callback(content)
		}
		return gerror.Newf("Acm listen empty message %s", dataID)
	})
}

// AddConfig 调用者注入监听
func AddConfig(dataID string, callback ParseContent, ignore ...bool) {
	cfg := &config{
		DataID:          dataID,
		Callback:        callback,
		IgnoreInitError: len(ignore) > 0 && ignore[0],
	}
	if ready {
		initConfig(cfg)
	} else {
		queue[dataID] = cfg
	}
}

// acmOnChange 定义监听acm变化回调函数
type acmOnChange func(dataID, data string) error

type consul struct {
	kv      *consulapi.Client
	listens map[string]acmOnChange
	changed map[string]uint64
	mu      sync.Mutex
}

func (serv *consul) init() {
	serv.mu.Lock()
	defer serv.mu.Unlock()
	if serv.kv == nil {
		cfg := &discoverConfig{}
		err := g.Cfg().GetStruct("rpc.discover", cfg)
		if err != nil {
			panic(err)
		}
		kv, err := consulapi.NewClient(&consulapi.Config{
			Address: cfg.Addr[0],
		})
		if err != nil {
			panic(err)
		}
		serv.kv = kv
	}
}
func (serv *consul) Get(dataID string) (string, error) {
	serv.init()
	key := fmt.Sprintf("%s/%s", prefix, dataID)
	pair, meta, err := serv.kv.KV().Get(key, &consulapi.QueryOptions{})
	if err != nil {
		g.Log().Println("acm get", dataID, "error", err)
		return "", err
	}
	if pair == nil {
		g.Log().Println("acm get", dataID, "not found")
		return "", gerror.Newf("key not found %s", key)
	}
	g.Log().Println("acm get ok", dataID, meta.LastIndex)
	serv.mu.Lock()
	serv.changed[key] = meta.LastIndex
	serv.mu.Unlock()
	return string(pair.Value), nil
}

func (serv *consul) GetJSON(dataID string, dataPtr interface{}) error {
	content, err := serv.Get(dataID)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(content), dataPtr)
}

func (serv *consul) GetYaml(dataID string, dataPtr interface{}) error {
	content, err := serv.Get(dataID)
	if err != nil {
		return err
	}
	return yaml.Unmarshal([]byte(content), dataPtr)
}

func (serv *consul) Listen(dataID string, onChange acmOnChange) {
	g.Log().Println("acm listen", dataID)
	serv.mu.Lock()
	defer serv.mu.Unlock()
	key := fmt.Sprintf("%s/%s", prefix, dataID)
	serv.listens[key] = onChange
}

func (serv *consul) Auto() {
	serv.init()
	client := serv.kv.KV()
	var waitIndex uint64 = 0
	for {
		pairs, meta, err := client.List(fmt.Sprintf("%s/", prefix), &consulapi.QueryOptions{
			WaitIndex: waitIndex,
			WaitTime:  time.Second * 60,
		})
		if err != nil || pairs == nil || len(pairs) == 0 {
			g.Log().Println("acm listen error", err)
			time.Sleep(time.Second)
			continue
		}
		serv.mu.Lock()
		for _, pair := range pairs {
			key := pair.Key
			if callback, ok := serv.listens[key]; ok {
				if preIndex, ok := serv.changed[key]; !ok || pair.ModifyIndex > preIndex {
					g.Log().Println("acm key changed", key, pair.ModifyIndex, preIndex)
					serv.changed[key] = pair.ModifyIndex
					err := callback(key, string(pair.Value))
					if err != nil {
						g.Log().Error("Acm auto listener callback error", err)
					}
				}
			}
		}
		waitIndex = meta.LastIndex
		serv.mu.Unlock()
	}
}
