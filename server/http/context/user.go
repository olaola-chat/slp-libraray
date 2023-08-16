package context

import (
	"context"
	"strings"

	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/text/gregex"
	"github.com/syyongx/php2go"

	"github.com/olaola-chat/rbp-library/tool"
)

var ContextSrv = new(contextService)

type contextService struct{}

func (s *contextService) GetUserCtx(ctx context.Context) *ContextUser {
	value := ctx.Value(ContextUserKey)
	if value == nil {
		return nil
	}
	if localCtx, ok := value.(*ContextUser); ok {
		return localCtx
	}
	return nil
}

// ContextUser 在请求上下文中的用户信息
type ContextUser struct {
	UID               uint32
	AppID             uint8 //用户对应的APPID
	Salt              string
	Platform          string //用户平台
	Time              uint32 //令牌生成时间
	Agent             string //用户的Agent
	Channel           string //用户所属渠道
	Package           string //当前请求的包名
	Language          string //用户的原始语言
	Area              string //用户的原始地区
	NativeVersion     uint32 //客户端版本号 ip2long
	NativeMainVersion uint32 //build号为0的版本
	JsVersion         uint32 //已经无用
	Mac               string
	DeviceName        string
	Did               string
	IsSimulator       bool //是否模拟器
	IsRoot            bool //是否Root设备
}

// IsLogined 判断当前用户是否登录
func (ctx *ContextUser) IsLogined() bool {
	return ctx.UID > 0 && ctx.AppID > 0 && ctx.Time > 0
}

// 定义不同APP
const (
	AppUnknown       uint8 = 0  //未备案
	AppRainbowPlanet uint8 = 88 //彩虹星球
)

const (
	PkgAppRainbowPlanetIos     = "com.im.duck.ios"
	PkgAppRainbowPlanetAndroid = "com.im.android.rbp"
)

var packageToAppID map[string]uint8 = map[string]uint8{
	PkgAppRainbowPlanetIos:     AppRainbowPlanet, //彩虹星球IOS
	PkgAppRainbowPlanetAndroid: AppRainbowPlanet, //彩虹星球android
}

// GetAppID 把客户端的包名转成对应的AppID
func GetAppID(packageName string) uint8 {
	if appID, ok := packageToAppID[packageName]; ok {
		return appID
	}
	return AppUnknown
}

// NewContextUserFromRequest 从http请求对象中获取用户的基本信息
func NewContextUserFromRequest(r *ghttp.Request) *ContextUser {
	user := &ContextUser{
		Package: r.GetString("package"),
	}
	user.AppID = GetAppID(user.Package)
	agent := r.GetHeader("User-Agent")
	if len(agent) == 0 {
		//这是小程序
		//todo... 需要确认究竟传递的是啥
		agent = r.GetHeader("U-A")
	}
	if len(agent) > 0 {
		referer := r.GetHeader("Referer")
		match, err := gregex.MatchString(`(?i)Xs (ios|android|pc|win32) V([0-9\.]+) \/ Js V([0-9\.]+)`, agent)
		if err == nil && len(match) > 0 && tool.IP.IsIPV4(match[2]) && tool.IP.IsIPV4(match[3]) {
			user.Agent = match[1] //ios | android | pc | win32
			user.NativeVersion = php2go.IP2long(match[2])
			user.JsVersion = php2go.IP2long(match[3])

			ipArray := strings.Split(match[2], ".")
			ipArray[3] = "0"
			user.NativeMainVersion = php2go.IP2long(strings.Join(ipArray, "."))
		} else if len(referer) > 0 && (strings.Contains(referer, "/pc") || strings.Contains(referer, "/stweb")) {
			user.Agent = "pc"
		}
	}

	userTag := r.GetHeader("User-Tag")
	if len(userTag) > 0 && userTag != "00000000-0000-0000-0000-000000000000" {
		user.Mac = userTag
	}

	user.DeviceName = r.GetHeader("User-Model")
	user.Channel = r.GetHeader("User-Channel")
	user.Did = r.GetHeader("User-Did")

	language := r.GetHeader("Accept-Language")
	if len(language) > 0 {
		//zh-CN,zh;q=0.9
		//todo... 优化，判断本地有哪些语言
		langArray := strings.Split(language, ",")
		user.Language = strings.ToLower(langArray[0])
	}
	language = r.GetHeader("User-Langauge")
	if len(language) > 0 {
		user.Language = strings.Trim(strings.ToLower(language), " ")
	} else {
		language = r.GetQueryString("lang")
		if len(language) > 0 {
			user.Language = strings.ToLower(language)
		}
	}
	user.IsSimulator = func() bool {
		if r.GetHeader("User-Issimulator") == "true" {
			return true
		}
		return tool.Device.IsEmulator(r.GetString("emulatorInfo"))
	}()
	user.IsRoot = func() bool {
		if r.GetHeader("User-Isroot") == "true" {
			return true
		}
		result := r.GetInt("isRooted", -1)
		if result < 0 {
			return result > 0
		}
		return false
	}()
	return user
}
