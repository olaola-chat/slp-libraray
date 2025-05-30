package main

import (
	"github.com/gogf/gf/frame/g"
	"github.com/olaola-chat/slp-library/acm"
	"time"
)

func main() {

	_ = acm.GetAcm().ListenKey("testkey", func(value string) error {
		g.Log().Info(value)
		return nil
	})

	_ = acm.GetAcm().ListenDir("testdir", func(kvs map[string]string) error {
		g.Log().Info(kvs)
		return nil
	})

	time.Sleep(100000 * time.Second)
}
