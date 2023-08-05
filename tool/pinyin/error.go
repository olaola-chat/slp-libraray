package pinyin

import "errors"

var (
	//ErrInitialize 初始化错误，一般是文件找不到时
	ErrInitialize = errors.New("not yet initialized")
)
