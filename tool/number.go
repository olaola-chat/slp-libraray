package tool

import (
	"github.com/gogf/gf/util/gconv"
	"strings"
)

// Path 单例selfpath，并导出
var NumberFormat = &numberFormat{}

type numberFormat struct{}

// 99,999,999.09
func (*numberFormat) ScienceFormat(num int) (value float32, unit string) {
	switch {
	case num < 1000:
		value = gconv.Float32(num)
		unit = ""
	case num < 1000000:
		value = gconv.Float32(num) / 1000
		unit = "K"
	case num < 1000000000:
		value = gconv.Float32(num) / 1000000
		unit = "M"
	default:
		value = gconv.Float32(num) / 1000000
		unit = "M"
	}
	return
}

// 格式化保留小数点相应位数
func (*numberFormat) DecimalPoint(value float32, radixNum int) float32 {
	str := gconv.String(value)
	strS := strings.Split(str, ".")
	var left, right string
	if len(strS) == 1 {
		left = strS[0]
		right = "0"
	} else {
		left = strS[0]
		right = strS[1]
	}
	rightVal := make([]byte, 0)
	rightLen := len(right)
	for i := 0; i < radixNum; i++ {
		if rightLen-1 < i {
			rightVal = append(rightVal, right[i])
		} else {
			rightVal = append(rightVal, '0')
		}
	}
	if len(rightVal) > 0 {
		str = left + "." + string(rightVal)
	} else {
		str = left
	}
	return gconv.Float32(str)
}

// 格式化成科学计数 10,000.989
func (*numberFormat) DecimalFormat(value float32, rightNum int) string {
	if value == 0 {
		return "0"
	}
	var left, right, str string
	str = gconv.String(value)
	strS := strings.Split(str, ".")
	if len(strS) == 1 {
		left = strS[0]
		right = "0"
	} else {
		left = strS[0]
		right = strS[1]
	}

	rightVal := make([]byte, 0)
	leftLen := len(left)
	base := ""
	if leftLen > 3 {
		start := leftLen % 3
		base = left[0:start]
		for start < leftLen-1 {
			base = base + "," + left[start:start+3]
			start = start + 3
		}
	} else {
		base = left
	}

	rightLen := len(right)
	for i := 0; i < rightNum; i++ {
		if i > rightLen-1 {
			rightVal = append(rightVal, '0')
		} else {
			rightVal = append(rightVal, right[i])
		}
	}
	str = base + "." + string(rightVal)
	return str
}
