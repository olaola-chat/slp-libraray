package binlog

import (
	"context"
	"encoding/json"
)

const (
	CanalWrite  = "INSERT"
	CanalDelete = "DELETE"
	CanalUpdate = "UPDATE"
)

// CanalJSON 定义DTS的数据结构
type CanalJSON struct {
	Data      []map[string]string `json:"data"`
	Database  string              `json:"database"`
	Table     string              `json:"table"`
	Es        int64               `json:"es"`        //操作在源库的执行时间，13位Unix时间戳，单位为毫秒
	Ts        int64               `json:"ts"`        //操作写入到目标库的时间，13位Unix时间戳，单位为毫秒
	Op        string              `json:"type"`      //操作的类型，比如DELETE、UPDATE、INSERT
	ID        int64               `json:"id"`        //操作的序列号
	IsDdl     bool                `json:"isDdl"`     //是否是DDL操作
	MysqlType map[string]string   `json:"mysqlType"` //字段的数据类型
	Old       []map[string]string `json:"old"`       //变更前的数据
	PkNames   []string            `json:"pkNames"`   //主键名称
	SQL       string              `json:"sql"`       //SQL语句
	SQLType   map[string]int      `json:"sqlType"`   //经转换处理后的字段类型
}

func ParseCanalJSON(data []byte) (*CanalJSON, error) {
	value := &CanalJSON{}
	err := json.Unmarshal(data, value)
	if err == nil {
		//兼容
		if value.Op == CanalDelete && len(value.Old) == 0 {
			value.Old = value.Data
		}
	}
	return value, err
}

type Callback interface {
	New() interface{}
	Inserted(ctx context.Context, val []interface{}, res *CanalJSON) error
	Updated(ctx context.Context, val []interface{}, res *CanalJSON) error
	Deleted(ctx context.Context, val []interface{}, res *CanalJSON) error
}

type BinlogCallback interface {
	New() interface{}
	Inserted(ctx context.Context, val interface{}) error
	Updated(ctx context.Context, val, old interface{}) error
	Deleted(ctx context.Context, val interface{}) error
}
