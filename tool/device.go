package tool

import (
	"encoding/json"
	"strings"
)

// Date 单例Date并导出
var Device = &device{}

type DeviceResult int8

const (
	DeviceMaybeEmulator DeviceResult = 0 //可能是模拟器
	DeviceIsEmulator    DeviceResult = 1 //模拟器
	DeviceMaybeReal     DeviceResult = 2 //可能是真机
)

type deviceInfo struct {
	IsIOSSimulator     int    `json:"isIOSSimulator"`
	Hardware           string `json:"hardware"`
	Flavor             string `json:"flavor"`
	Model              string `json:"model"`
	Manufacturer       string `json:"manufacturer"`
	Board              string `json:"board"`
	Platform           string `json:"platform"`
	BaseBand           string `json:"baseBand"`
	CgroupResult       string `json:"cgroupResult"`
	SensorNumber       int    `json:"sensorNumber"`
	UserAppNumber      int    `json:"userAppNumber"`
	SupportCameraFlash bool   `json:"supportCameraFlash"`
	SupportCamera      bool   `json:"supportCamera"`
	SupportBluetooth   bool   `json:"supportBluetooth"`
	HasLightSensor     bool   `json:"hasLightSensor"`
}

type device struct {
}

// IsEmulator 特征参数-进程组信息
func (d *device) IsEmulator(jsonStr string) bool {
	info := deviceInfo{IsIOSSimulator: -1}
	err := json.Unmarshal([]byte(jsonStr), &info)
	if err != nil {
		//解码错误应该判断为模拟器，否则参数意义何在
		return true
	}
	if info.IsIOSSimulator >= 0 {
		return info.IsIOSSimulator == 1
	}

	res := []DeviceResult{
		d.checkFeaturesByHardware(info.Hardware),
		d.checkFeaturesByFlavor(info.Flavor),
		d.checkFeaturesByModel(info.Model),
		d.checkFeaturesByManufacturer(info.Manufacturer),
		d.checkFeaturesByBoard(info.Board),
		d.checkFeaturesByPlatform(info.Platform),
		d.checkFeaturesByBaseBand(info.BaseBand),
	}
	suspectCount := 0
	for _, r := range res {
		if r == DeviceIsEmulator {
			return true
		}
		suspectCount += int(r)
	}

	//检测传感器数量
	if info.SensorNumber <= 7 {
		suspectCount++
	}
	//检测已安装第三方应用数量
	if info.UserAppNumber <= 5 {
		suspectCount++
	}
	//检测是否支持闪光灯
	if !info.SupportCameraFlash {
		suspectCount++
	}
	//检测是否支持相机
	if !info.SupportCamera {
		suspectCount++
	}
	//检测是否支持蓝牙
	if !info.SupportBluetooth {
		suspectCount++
	}
	//检测光线传感器
	if !info.HasLightSensor {
		suspectCount++
	}
	//检测进程组信息
	if d.checkFeaturesByCgroup(info.CgroupResult) == DeviceMaybeEmulator {
		suspectCount++
	}
	return suspectCount > 3
}

// checkFeaturesByCgroup 特征参数-进程组信息
func (*device) checkFeaturesByCgroup(value string) DeviceResult {
	if len(value) == 0 {
		return DeviceMaybeEmulator
	}
	return DeviceMaybeReal
}

// checkFeaturesByBaseBand 特征参数-基带信息
func (*device) checkFeaturesByBaseBand(value string) DeviceResult {
	if len(value) == 0 {
		return DeviceMaybeEmulator
	}
	value = strings.ToLower(value)
	if strings.Contains(value, "1.0.0.0") {
		return DeviceIsEmulator
	}
	return DeviceMaybeReal
}

// checkFeaturesByPlatform 特征参数-主板平台
func (*device) checkFeaturesByPlatform(value string) DeviceResult {
	if len(value) == 0 {
		return DeviceMaybeEmulator
	}
	value = strings.ToLower(value)
	if strings.Contains(value, "android") {
		return DeviceIsEmulator
	}
	return DeviceMaybeReal
}

// checkFeaturesByBoard 特征参数-主板名称
func (*device) checkFeaturesByBoard(value string) DeviceResult {
	if len(value) == 0 {
		return DeviceMaybeEmulator
	}
	value = strings.ToLower(value)
	if strings.Contains(value, "android") || strings.Contains(value, "goldfish") {
		return DeviceIsEmulator
	}
	return DeviceMaybeReal
}

// checkFeaturesByManufacturer 特征参数-硬件制造商
func (*device) checkFeaturesByManufacturer(value string) DeviceResult {
	if len(value) == 0 {
		return DeviceMaybeEmulator
	}
	value = strings.ToLower(value)
	if strings.Contains(value, "genymotion") || strings.Contains(value, "netease") {
		return DeviceIsEmulator
	}
	return DeviceMaybeReal
}

// checkFeaturesByModel 特征参数-设备型号
func (*device) checkFeaturesByModel(value string) DeviceResult {
	if len(value) == 0 {
		return DeviceMaybeEmulator
	}
	value = strings.ToLower(value)
	if strings.Contains(value, "google_sdk") ||
		strings.Contains(value, "emulator") ||
		strings.Contains(value, "android sdk built for x86") {
		return DeviceIsEmulator
	}
	return DeviceMaybeReal
}

// checkFeaturesByFlavor 特征参数-渠道
func (*device) checkFeaturesByFlavor(value string) DeviceResult {
	if len(value) == 0 {
		return DeviceMaybeEmulator
	}
	value = strings.ToLower(value)
	if strings.Contains(value, "vbox") || strings.Contains(value, "sdk_gphone") {
		return DeviceIsEmulator
	}
	return DeviceMaybeReal
}

// checkFeaturesByHardware 特征参数-硬件名称
func (*device) checkFeaturesByHardware(value string) DeviceResult {
	if len(value) == 0 {
		return DeviceMaybeEmulator
	}
	switch strings.ToLower(value) {
	case "ttvm", "nox", "cancro", "intel", "vbox", "vbox86", "android_x86":
		//天天模拟器
		//夜神模拟器
		//网易MUMU模拟器
		//逍遥模拟器
		//...
		//腾讯手游助手
		//雷电模拟器
		return DeviceIsEmulator
	}
	return DeviceMaybeReal
}
