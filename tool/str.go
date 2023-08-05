package tool

import (
	"bytes"
	"github.com/gogf/gf/text/gstr"
	"github.com/gogf/gf/util/gconv"
	"strconv"
	"strings"
	"unicode"
)

// Str 导出str对象
var Str = &str{}

type str struct{}

// UnescapeUnicode Unicode 变成字符串
func (*str) UnescapeUnicode(raw []byte) ([]byte, error) {
	str, err := strconv.Unquote(strings.Replace(strconv.Quote(string(raw)), `\\u`, `\u`, -1))
	if err != nil {
		return nil, err
	}
	return []byte(str), nil
}

// EscapeUnicode 字符串变成\u Unicode
func (*str) EscapeUnicode(raw string) string {
	textQuoted := strconv.QuoteToASCII(raw)
	return textQuoted[1 : len(textQuoted)-1]
}

func (*str) CleanString(word string) string {
	buffer := &bytes.Buffer{}
	for _, r := range word {
		if unicode.IsLetter(r) {
			if r >= 65345 && r < 65371 {
				continue
			} else if r >= 65313 && r < 65339 {
				continue
			} else if r < 255 {
				continue
			} else {
				buffer.WriteRune(r)
			}
		} else if unicode.IsSymbol(r) {
			if unicode.IsPrint(r) || unicode.IsGraphic(r) {
				continue
			}
			buffer.WriteRune(r)
		}
	}
	return buffer.String()
}

func (*str) ToUint32ArrayWithLimit(str string, limit int) []uint32 {
	if str == "" {
		return nil
	}
	list := gstr.Explode(",", str)
	data := gconv.Uint32s(list)
	if len(data) > limit {
		data = data[:limit]
	}
	return data
}

// 拼接字符串
func (*str) StrJoin(connectStr string, str ...string) string {
	if len(str) == 0 {
		return ""
	}
	var strs strings.Builder
	for i, s := range str {
		strs.WriteString(s)
		if i < len(str)-1 && connectStr != "" {
			strs.WriteString(connectStr)
		}
	}
	return strs.String()
}

func (*str) FirstToUpper(str string) string {
	if len(str) < 1 {
		return ""
	}
	strArr := []rune(str)
	if strArr[0] >= 'a' && strArr[0] <= 'z' {
		strArr[0] -= 'a' - 'A'
	}
	return string(strArr)
}
