package tool

import (
	"reflect"
	"sort"
)

type ArraySort struct {
	Data      []map[string]interface{}
	Match     func(itemI, itemJ map[string]interface{}) bool //小于号比较为倒序
	OrderDesc bool
}

func (this ArraySort) Len() int {
	return len(this.Data)
}

func (this ArraySort) Swap(i, j int) {
	this.Data[i], this.Data[j] = this.Data[j], this.Data[i]
}

func (this ArraySort) Less(i, j int) bool {
	if this.OrderDesc {
		return !this.Match(this.Data[i], this.Data[j])
	}
	return this.Match(this.Data[i], this.Data[j])

}

// 数组排序
func Sort(data ArraySort) []map[string]interface{} {
	sort.Stable(data)
	return data.Data
}

// 转为关联数组
func Struct2Array(data interface{}, tagName string) map[string]interface{} {
	returnData := make(map[string]interface{})
	var typeInfo = reflect.TypeOf(&data)
	var valInfo = reflect.ValueOf(&data)
	num := typeInfo.NumField()
	for i := 0; i < num; i++ {
		tagValue := typeInfo.Field(i)
		key := tagValue.Tag.Get(tagName)
		if tagValue.Tag.Get(tagName) == "" {
			continue
		}
		returnData[key] = valInfo.Field(i).Interface()
	}
	return returnData
}
