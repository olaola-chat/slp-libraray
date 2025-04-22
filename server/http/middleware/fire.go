package middleware

import (
	"encoding/binary"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/frame/g"

	"github.com/olaola-chat/slp-library/tool"

	mapset "github.com/deckarep/golang-set"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/util/gconv"
	"github.com/syyongx/php2go"
)

var _fire *firewall
var fireOnce sync.Once

func Fire(r *ghttp.Request) {
	fireOnce.Do(func() {
		_fire = &firewall{}
		_fire.init()
		go _fire.tick()
	})
	if !_fire.add(r) {
		r.Response.Status = http.StatusNotAcceptable
		r.Exit()
		return
	}
	r.Middleware.Next()
}

type firewall struct {
	fields       []string
	data         []*map[uint32]uint32
	replay       []mapset.Set
	mu           sync.Mutex
	second       int
	blockIP      map[string]bool
	dangerRegexp *regexp.Regexp
}

const maxReqSecond = 10

// http server启动前执行
// 新启动一个线程用于接收数据，接收的数据都向这个线程里放，避免多线程之间的各种竞争

func (fire *firewall) init() {
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
	fire.fields = keys
	fire.data = make([]*map[uint32]uint32, 60)
	fire.replay = make([]mapset.Set, 30)
	fire.second = time.Now().Second()
	fire.blockIP = make(map[string]bool)

	for i := 0; i < 60; i++ {
		fire.data[i] = &map[uint32]uint32{}
	}
	for i := 0; i < 30; i++ {
		//生成一个线程安全的set集合
		fire.replay[i] = mapset.NewSet()
	}
	str := `(?i)(\b(shell|exec|select|update|delete|insert|trancate|char|chr|substr|ascii|declare|master|drop|execute)\b)`
	var err error
	fire.dangerRegexp, err = regexp.Compile(str)
	if err != nil {
		g.Log().Error("danger Regexp error ", err)
		// panic(err.Error())
	}
}

func (fire *firewall) tick() {
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

func (fire *firewall) add(r *ghttp.Request) bool {
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
			if num+1 > maxReqSecond {
				return false
			}
		} else {
			res[ip] = 1
		}
	}
	//检测请求中是否含有危险的词汇
	if fire.dangerRegexp != nil && fire.dangerRegexp.Match([]byte(r.URL.RawQuery)) {
		return false
	}
	if !r.IsFileRequest() && fire.dangerRegexp != nil && fire.dangerRegexp.Match(r.GetBody()) {
		return false
	}

	return fire.sign(r)
}

func (*firewall) hashTime(t time.Time) int {
	return t.Minute() / 2
}

// 验证请求加密信息，防止重放，虽然没有完全意义，但总比没有好
// 前端会根据用户ip hash来分发，这样同一个人在同一台机器上
// 要考虑同一个用户有多端登录
// 客户端时间差的太多的(分析下现在日志，缩小时间判断差值...)
// todo... 重放攻击
func (fire *firewall) sign(r *ghttp.Request) bool {
	//客户端时间差超过28分钟，也要被屏蔽
	now := time.Now().Unix()
	timestamp := gconv.Int64(r.GetQuery("_timestamp"))
	if now-timestamp > 28*60 {
		return false
	}
	//客户端签名不合法，屏蔽
	querySign := r.GetQuery("_sign")
	signVer := r.GetQueryInt("_sign_ver")
	res := []string{}
	for _, key := range fire.fields {
		res = append(res, fmt.Sprintf("%s=%s", key, r.GetQuery(key)))
	}
	salt := "!rilegoule#"
	if signVer == 1 {
		salt = "!caihongmeng#"
	}
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
	return fire.replay[idx].Add(querySign)
	//先关闭验证，直接返回通过
	// return true
}
