package region

import (
	"bufio"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gogf/gf/errors/gerror"
	"github.com/gogf/gf/util/gconv"
)

// CountryInfo 定义一组国家信息
type CountryInfo struct {
	NameCn    string `json:"name_cn"`
	Alpha2    string `json:"alpha_2"`
	Alpha3    string `json:"alpha_3"`
	PhoneCode int32  `json:"phone_code"`
}

// Copy 复制当前数据
func (s *CountryInfo) Copy() CountryInfo {
	return CountryInfo{
		NameCn:    s.NameCn,
		Alpha2:    s.Alpha2,
		Alpha3:    s.Alpha3,
		PhoneCode: s.PhoneCode,
	}
}

// ErrorNotFound 定义不存在次国家信息的error
var ErrorNotFound = gerror.New("not found")
var countryDataFile string
var emptyCountryInfo = CountryInfo{}

// FilerType 定义查询类型
type FilerType int8

const (
	//FilerTypeNameCn 按照国家查询
	FilerTypeNameCn FilerType = iota
	//FilerTypeAlpha2 按照alpha2查询
	FilerTypeAlpha2
	//FilerTypeAlpha3 按照alpha3查询
	FilerTypeAlpha3
)

func init() {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}
	countryDataFile = path.Join(dir, "config", "country.csv")
}

// Info 定义一个查询单例
var Info = &info{
	data: make([]*CountryInfo, 0),
}

type info struct {
	initialized bool
	mu          sync.Mutex
	data        []*CountryInfo
	nameTo      map[string]*CountryInfo
	alpha2To    map[string]*CountryInfo
	alpha3To    map[string]*CountryInfo
}

func (serv *info) initWrap() error {
	if serv.initialized {
		return nil
	}
	return serv.init()
}

func (serv *info) init() error {
	serv.mu.Lock()
	defer serv.mu.Unlock()

	if serv.initialized {
		return nil
	}
	fp, err := os.Open(countryDataFile)
	if err != nil {
		return err
	}
	defer fp.Close()

	r := bufio.NewReader(fp)
	for {
		a, _, c := r.ReadLine()
		if c == io.EOF {
			break
		}
		lines := strings.Split(string(a), ",")
		if len(lines) == 4 {
			cc := &CountryInfo{
				NameCn:    lines[2],
				Alpha2:    lines[0],
				Alpha3:    lines[1],
				PhoneCode: gconv.Int32(lines[3]),
			}
			serv.data = append(serv.data, cc)
			serv.nameTo[cc.NameCn] = cc
			serv.alpha2To[cc.Alpha2] = cc
			serv.alpha3To[cc.Alpha3] = cc
		}
	}
	serv.initialized = true
	return nil
}

func (serv *info) By(filter FilerType, value string) (CountryInfo, error) {
	err := serv.initWrap()
	if err != nil {
		return emptyCountryInfo, err
	}
	var data map[string]*CountryInfo
	switch filter {
	case FilerTypeNameCn:
		data = serv.nameTo
	case FilerTypeAlpha2:
		data = serv.alpha2To
	case FilerTypeAlpha3:
		data = serv.alpha3To
	}
	if cc, ok := data[value]; ok {
		return cc.Copy(), nil
	}
	return emptyCountryInfo, ErrorNotFound
}
