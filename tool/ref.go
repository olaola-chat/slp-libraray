package tool

import (
	"reflect"
	"strconv"
)

// Ref 单例selfref，并导出
var Refe = &refe{}

type refe struct{}

func (r *refe) GetReflectInt(value interface{}) int64 {
	if value == nil {
		return 0
	}

	vtype := reflect.TypeOf(value).String()
	if vtype == "string" {
		val, err := strconv.ParseInt(reflect.ValueOf(value).String(), 10, 64)
		if err == nil {
			return val
		}
		return 0
	}

	if vtype == "int" || vtype == "int64" || vtype == "int32" {
		return reflect.ValueOf(value).Int()
	}

	if vtype == "float" || vtype == "float32" || vtype == "float64" {
		return int64(float64(reflect.ValueOf(value).Float()))
	}

	return 0
}

func (r *refe) GetReflectString(value interface{}) string {
	if value == nil {
		return ""
	}

	vtype := reflect.TypeOf(value).String()
	if vtype == "string" {
		return reflect.ValueOf(value).String()
	}

	return ""
}
