package tool

import (
	"fmt"
	"regexp"
	"strings"
)

var Img = &img{}

type img struct{}

const (
	CDN_IMG_PROXY_DOMAIN = "https://image.caihongxq.com"
)

func (i *img) AppendCdnHost(ul string) string {

	if ok, _ := regexp.Match("^http", []byte(ul)); ok {
		return ul
	}

	return fmt.Sprintf("%v/%v", CDN_IMG_PROXY_DOMAIN, strings.Trim(ul, "/"))

}
