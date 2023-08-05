package loghook

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"

	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/glog"
	"github.com/gogf/gf/text/gregex"
)

// NewLogWriter 自定义日志格式，从配置文件中读取logger配置
func NewLogWriter(serverName string) *LogWriter {
	logger := glog.New()
	m := g.Config().GetMap("logger")
	if len(m) > 0 {
		if err := logger.SetConfigWithMap(m); err != nil {
			panic(err)
		}
	}

	return &LogWriter{
		serverName: serverName,
		logger:     logger,
	}
}

type LogWriter struct {
	serverName string
	logger     *glog.Logger
}

// Write ERROR以上级别日志，加上堆栈信息
func (w *LogWriter) Write(p []byte) (n int, err error) {
	str := string(p)
	var buffer bytes.Buffer
	buffer.WriteString("[" + time.Now().Format("2006-01-02 15:04:05") + "]" + "[" + w.serverName + "]")
	if !gregex.IsMatchString(`WARN|NOTI|INFO|ERRO|CRIT|PANI|FATA`, str) {
		buffer.WriteString("[INFO]")
		buffer.WriteByte(' ')
	}
	buffer.WriteString(strings.Trim(string(p), "\n"))
	if gregex.IsMatchString(`ERRO|CRIT|PANI|FATA`, str) {
		buffer.WriteString("，Stack: ")
		bs, _ := json.Marshal(strings.Split(strings.Trim(w.logger.GetStack(1), "\n"), "\n"))
		buffer.Write(bs)
	}
	return w.logger.Write(buffer.Bytes())
}

func (w *LogWriter) Output(calldepth int, str string) error {
	var buffer bytes.Buffer
	buffer.WriteString("[" + time.Now().Format("2006-01-02 15:04:05") + "]" + "[" + w.serverName + "]")
	buffer.WriteString(strings.Trim(str, "\n"))
	if gregex.IsMatchString(`ERR|CRIT|PANI|FATA`, str) {
		buffer.WriteString("，Stack: ")
		bs, _ := json.Marshal(strings.Split(strings.Trim(w.logger.GetStack(1), "\n"), "\n"))
		buffer.Write(bs)
	}
	_, err := w.logger.Write(buffer.Bytes())
	return err
}
