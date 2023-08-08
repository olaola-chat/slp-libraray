package http

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/olaola-chat/rbp-library/acm"
	"github.com/olaola-chat/rbp-library/env"
	"github.com/olaola-chat/rbp-library/tool"

	mapset "github.com/deckarep/golang-set"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/util/gconv"
	"github.com/syyongx/php2go"
)

// Firewall 导出单例，用于全局计算
// ip 频率限制
// 防止回放，防止参数篡改
// 防止来自机房的请求...
// ip 屏蔽，country 屏蔽
// Sql 注入...
// 定时上报topN Ip
// 接收任务，屏蔽指定请求
var Firewall *firewallService
var dangerRegexp *regexp.Regexp

const safeReqSecond = 10

func init() {
	keys := sort.StringSlice{
		"package",
		"_ipv",
		"_platform",
		"_index",
		"_model",
		"_timestamp",
		"format",
	}
	sort.Sort(keys)
	Firewall = &firewallService{
		fields: keys,
		data:   make([]*map[uint32]uint32, 60),
		replay: make([]mapset.Set, 30),
		config: &acmFirewallConfig{
			BlockIP:      []string{},
			MaxReqSecond: safeReqSecond,
		},
		second:  time.Now().Second(),
		blockIP: make(map[string]bool),
	}
	for i := 0; i < 60; i++ {
		Firewall.data[i] = &map[uint32]uint32{}
	}
	for i := 0; i < 30; i++ {
		//生成一个线程安全的set集合
		Firewall.replay[i] = mapset.NewSet()
	}
	str := `(?i)(\b(shell|exec|select|update|delete|insert|trancate|char|chr|substr|ascii|declare|master|drop|execute)\b)`
	var err error
	dangerRegexp, err = regexp.Compile(str)
	if err != nil {
		panic(err.Error())
	}

	acm.GetAcm().ListenKey("config.firewall", func(content string) error {
		v := &acmFirewallConfig{}
		err := json.Unmarshal([]byte(content), v)
		if err != nil {
			g.Log().Println("acm parse cofig.firewall error", err)
			return err
		}
		block := make(map[string]bool)
		for _, ip := range v.BlockIP {
			block[ip] = true
		}
		Firewall.mu.Lock()
		Firewall.blockIP = block
		Firewall.config = v
		Firewall.mu.Unlock()
		return nil
	})
}

type acmFirewallConfig struct {
	BlockIP      []string `json:"block_ip"`
	MaxReqSecond uint32   `json:"max_req_second"`
}

type firewallService struct {
	fields  []string
	data    []*map[uint32]uint32
	replay  []mapset.Set
	config  *acmFirewallConfig
	mu      sync.Mutex
	second  int
	blockIP map[string]bool
}

// http server启动前执行
// 新启动一个线程用于接收数据，接收的数据都向这个线程里放，避免多线程之间的各种竞争
func (fire *firewallService) Init() {
	counter := 0
	ticker := time.NewTicker(time.Second)
	for range ticker.C {
		fire.mu.Lock()
		fire.second = time.Now().Second()
		fire.data[fire.second] = &map[uint32]uint32{}
		fire.mu.Unlock()

		//每分钟清除未来时间分片的set集合
		counter++
		if counter == 60 {
			counter = 0
			idx := fire.hashTime(time.Now().Add(time.Minute * 2))
			fire.replay[idx].Clear()
		}
	}
}

func (fire *firewallService) Add(r *ghttp.Request) bool {
	addr := r.GetClientIp()
	if len(addr) > 0 && !tool.IP.IsLanIP(addr) {
		if _, ok := fire.blockIP[addr]; ok {
			return false
		}
		var ip uint32 = 0
		ipv := net.ParseIP(addr)
		if ipv != nil {
			if strings.Contains(addr, ".") { //ipv4
				ip = binary.BigEndian.Uint32(ipv.To4())
			} else if strings.Contains(addr, ":") { // ipv6
				ip = binary.BigEndian.Uint32(gconv.Bytes(addr))
			}
		}
		//todo... 优化
		fire.mu.Lock()
		defer fire.mu.Unlock()
		res := *fire.data[fire.second]
		if num, ok := res[ip]; ok {
			res[ip] = num + 1
			if fire.config.MaxReqSecond >= safeReqSecond && num+1 > fire.config.MaxReqSecond {
				return false
			}
		} else {
			res[ip] = 1
		}
	}
	//检测请求中是否含有危险的词汇
	if dangerRegexp.Match([]byte(r.URL.RawQuery)) {
		return false
	}
	if !r.IsFileRequest() && dangerRegexp.Match(r.GetBody()) {
		return false
	}

	return fire.sign(r)
}

func (*firewallService) hashTime(t time.Time) int {
	return t.Minute() / 2
}

// 验证请求加密信息，防止重放，虽然没有完全意义，但总比没有好
// 前端会根据用户ip hash来分发，这样同一个人在同一台机器上
// 要考虑同一个用户有多端登录
// 客户端时间差的太多的(分析下现在日志，缩小时间判断差值...)
// todo... 重放攻击
func (fire *firewallService) sign(r *ghttp.Request) bool {
	if env.IsDev {
		return true
	}
	//客户端时间差超过28分钟，也要被屏蔽
	now := time.Now().Unix()
	timestamp := gconv.Int64(r.GetQuery("_timestamp"))
	if now-timestamp > 28*60 {
		return false
	}
	//客户端签名不合法，屏蔽
	querySign := r.GetQuery("_sign")
	res := []string{}
	for _, key := range fire.fields {
		res = append(res, fmt.Sprintf("%s=%s", key, r.GetQuery(key)))
	}
	salt := "!rilegoule#"
	// if isMini {
	// 	salt = "!mini#"
	// }
	sign := php2go.Md5(strings.Join(res, "&") + salt)
	if sign != querySign {
		return false
	}
	//防止重放
	//把 timestamp 散列到30个set集合里
	//在此set集合里寻找sign是否存在，如果存在，则屏蔽
	//不存在则添加到此set
	//然后每一分钟，清除未来两分钟需要使用的那个集合
	//按照单个进程每秒2000qps，需要使用100MB+内存
	//客户端的请求时间能够和服务器校准，那么就会简单很多
	idx := fire.hashTime(time.Unix(timestamp, 0))
	fire.replay[idx].Add(querySign)
	//先关闭验证，直接返回通过
	return true
}
