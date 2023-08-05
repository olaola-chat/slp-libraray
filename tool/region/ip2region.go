package region

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"sync"
)

//更改为线程安全memory search

const (
	//IndexBlockLength 宏定义
	IndexBlockLength = 12
)

// IP2Region ip到地址的查询
type IP2Region struct {
	firstIndexPtr int64
	lastIndexPtr  int64
	totalBlocks   int64
	dbBinStr      []byte
	init          bool
	mu            sync.RWMutex
}

// IPInfo 单个IP对应的地区信息
type IPInfo struct {
	Country  string
	Region   string
	Province string
	City     string
	ISP      string
}

// String 可视化输出
func (ip IPInfo) String() string {
	return ip.Country + "|" + ip.Region + "|" + ip.Province + "|" + ip.City + "|" + ip.ISP
}

func getIPInfo(line []byte) IPInfo {
	lineSlice := strings.Split(string(line), "|")
	ipInfo := IPInfo{}
	length := len(lineSlice)
	if length < 5 {
		for i := 0; i <= 5-length; i++ {
			lineSlice = append(lineSlice, "")
		}
	}

	ipInfo.Country = lineSlice[0]
	ipInfo.Region = lineSlice[1]
	ipInfo.Province = lineSlice[2]
	ipInfo.City = lineSlice[3]
	ipInfo.ISP = lineSlice[4]
	return ipInfo
}

// Initialized 返回此实例是否已经初始化
func (serv *IP2Region) Initialized() bool {
	return serv.init
}

// Init 初始化资源
func (serv *IP2Region) Init(file string) error {
	serv.mu.Lock()
	defer serv.mu.Unlock()
	if serv.init {
		return nil
	}
	var err error
	serv.dbBinStr, err = os.ReadFile(file)
	if err != nil {
		return err
	}
	serv.firstIndexPtr = getLong(serv.dbBinStr, 0)
	serv.lastIndexPtr = getLong(serv.dbBinStr, 4)
	serv.totalBlocks = (serv.lastIndexPtr-serv.firstIndexPtr)/IndexBlockLength + 1
	serv.init = true
	return nil
}

// MemorySearch 内存查找
func (serv *IP2Region) MemorySearch(ipv4 string) (ipInfo IPInfo, err error) {
	ipInfo = IPInfo{}

	ip, err := ip2long(ipv4)
	if err != nil {
		return ipInfo, err
	}

	h := serv.totalBlocks
	var dataPtr, l int64
	for l <= h {

		m := (l + h) >> 1
		p := serv.firstIndexPtr + m*IndexBlockLength
		sip := getLong(serv.dbBinStr, p)
		if ip < sip {
			h = m - 1
		} else {
			eip := getLong(serv.dbBinStr, p+4)
			if ip > eip {
				l = m + 1
			} else {
				dataPtr = getLong(serv.dbBinStr, p+8)
				break
			}
		}
	}
	if dataPtr == 0 {
		return ipInfo, errors.New("not found")
	}

	dataLen := ((dataPtr >> 24) & 0xFF)
	dataPtr = (dataPtr & 0x00FFFFFF)
	ipInfo = getIPInfo(serv.dbBinStr[(dataPtr)+4 : dataPtr+dataLen])
	return ipInfo, nil

}

func getLong(b []byte, offset int64) int64 {

	val := (int64(b[offset]) |
		int64(b[offset+1])<<8 |
		int64(b[offset+2])<<16 |
		int64(b[offset+3])<<24)

	return val

}

func ip2long(IPStr string) (int64, error) {
	bits := strings.Split(IPStr, ".")
	if len(bits) != 4 {
		return 0, errors.New("ip format error")
	}

	var sum int64
	for i, n := range bits {
		bit, _ := strconv.ParseInt(n, 10, 64)
		sum += bit << uint(24-8*i)
	}

	return sum, nil
}
