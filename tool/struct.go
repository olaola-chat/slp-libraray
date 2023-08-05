package tool

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/gogf/gf/errors/gerror"
)

// TagType 尝试判断结构体中有何种tag
type TagType int

const (
	//TagTypeUnknown 未知
	TagTypeUnknown TagType = iota
	//TagTypePb struct中包含pb tag
	TagTypePb
	//TagTypeOrm struct中包含orm tag
	TagTypeOrm
)

// Ref 用于model和pb数据的相互复制
var Ref = &ref{
	tag: make(map[string][]string),
}

type ref struct {
	tag map[string][]string //缓存pb, model数据的protobuf/orm tag映射
}

// Init 在启动时调用注册
func (serv *ref) Init(source interface{}) {
	err := serv.from(source)
	if err != nil {
		panic(err)
	}
}

// 获取source对象的tag=>name映射
func (serv *ref) from(source interface{}) error {
	vVal := reflect.ValueOf(source).Elem() //获取reflect.Type类型
	vTypeOfT := vVal.Type()
	key := fmt.Sprintf("%s/%s", vTypeOfT.PkgPath(), vTypeOfT.Name())

	tagType := TagTypeUnknown
	for i := 0; i < vVal.NumField(); i++ {
		tag := vTypeOfT.Field(i).Tag.Get("protobuf")
		if len(tag) > 0 {
			tagType = TagTypePb
			break
		}
		tag = vTypeOfT.Field(i).Tag.Get("orm")
		if len(tag) > 0 {
			tagType = TagTypeOrm
			break
		}
	}

	if tagType == TagTypeUnknown {
		return gerror.New("unknown struct data")
	}

	data := make([]string, vVal.NumField())
	for i := 0; i < vVal.NumField(); i++ {
		if tagType == TagTypePb {
			tag := vTypeOfT.Field(i).Tag.Get("protobuf")
			if len(tag) > 0 {
				s := strings.Index(tag, "name=")
				l := strings.Index(tag[s+5:], ",")
				dbName := tag[s+5 : s+5+l]
				data[i] = dbName
			}
		} else {
			tag := vTypeOfT.Field(i).Tag.Get("orm")
			if len(tag) > 0 {
				ts := strings.Split(tag, ",")
				dbName := ts[0]
				data[i] = dbName
			}
		}
	}
	serv.tag[key] = data
	return nil
}

func (serv *ref) getKey(source interface{}) string {
	vVal := reflect.ValueOf(source).Elem() //获取reflect.Type类型
	vTypeOfT := vVal.Type()
	return fmt.Sprintf("%s/%s", vTypeOfT.PkgPath(), vTypeOfT.Name())
}

// RetainField 保留结构体某些字段，其他字段充值为默认值
func (serv *ref) RetainFieldSlice(source interface{}, fields []string) {
	serv.RetainFieldMap(source, serv.SliceToMap(fields))
}

func (serv *ref) SliceToMap(fields []string) map[string]bool {
	fieldMap := make(map[string]bool)
	for _, field := range fields {
		fieldMap[field] = true
	}
	return fieldMap
}

func (serv *ref) getMapping(source interface{}) []string {
	key := serv.getKey(source)
	if mapping, ok := serv.tag[key]; ok {
		return mapping
	}
	panic(fmt.Errorf("error to find mapping from source %v", source))
}

func (serv *ref) RetainFieldMap(source interface{}, fieldMap map[string]bool) {
	if len(fieldMap) == 0 {
		return
	}
	mapping := serv.getMapping(source)
	vVal := reflect.ValueOf(source).Elem() //获取reflect.Type类型
	for i := 0; i < vVal.NumField(); i++ {
		value := vVal.Field(i)
		dbName := mapping[i]
		if _, ok := fieldMap[dbName]; !ok && value.CanSet() {
			value.Set(reflect.Zero(value.Type()))
		}
	}
}
