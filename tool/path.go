package tool

import (
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gogf/gf/os/genv"
)

// Path 单例selfpath，并导出
var Path = &selfpath{}

type selfpath struct{}

// ExecPath 获取当前可执行文件的路径
func (*selfpath) ExecPath() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}
	return dir
}

// ExecRootPath 获取当前项目的根目录
func (s *selfpath) ExecRootPath() string {
	dir := genv.Get("GF_GCFG_PATH")
	if len(dir) > 0 {
		return dir
	}
	dir = s.ExecPath()
	if strings.HasSuffix(dir, "bin") {
		return path.Dir(dir)
	}
	return dir
}

// GetFilePath 获取当前path.go所在路径
func (*selfpath) GetFilePath() string {
	_, filename, _, _ := runtime.Caller(1)
	return path.Dir(filename)
}
