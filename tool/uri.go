package tool

import (
	"strings"
)

var URI = &URIObj{
	prefixs: []string{"http://", "https://"},
}

type URIObj struct {
	prefixs []string
}

func (o *URIObj) Combine(host, path string) string {
	if host == "" || path == "" {
		return path
	}

	for _, v := range o.prefixs {
		if strings.HasPrefix(path, v) {
			return path
		}
	}

	sep := "/"
	if strings.HasSuffix(host, "/") ||
		strings.HasPrefix(path, "/") {
		sep = ""
	}

	return host + sep + path
}
