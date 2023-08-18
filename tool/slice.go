package tool

import (
	"fmt"
	"github.com/gogf/gf/container/gset"
	"github.com/gogf/gf/util/gconv"
	"math/rand"
	"sort"
)

// Str 导出str对象
var Slice = &slice{}

type slice struct{}

// InStringArray 判断字符串是否在string slice中
func (*slice) InStringArray(value string, array []string) bool {
	if array == nil {
		return false
	}
	length := len(array)
	for i := 0; i < length; i++ {
		if array[i] == value {
			return true
		}
	}
	return false
}

// 获取数组中的最大和最小值
func (*slice) GetMaxAndMin(arr []int) (int, int) {

	max := 0
	min := 0
	if len(arr) <= 0 {
		return max, min
	}

	for i := 0; i < len(arr); i++ {
		if arr[i] > max {
			max = arr[i]
		}
		if arr[i] < min {
			min = arr[i]
		}
	}
	return max, min
}

func (*slice) InUint32Array(value uint32, array []uint32) bool {
	if array == nil {
		return false
	}
	length := len(array)
	for i := 0; i < length; i++ {
		if array[i] == value {
			return true
		}
	}
	return false
}

func (*slice) InInt32Array(value int32, array []int32) bool {
	if array == nil {
		return false
	}
	length := len(array)
	for i := 0; i < length; i++ {
		if array[i] == value {
			return true
		}
	}
	return false
}

// 从arr中过滤掉filter中的数据
func (s *slice) FilterUint32Array(arr []uint32, filter []uint32) []uint32 {
	ret := make([]uint32, 0, len(arr))
	for _, v := range arr {
		if !s.InUint32Array(v, filter) {
			ret = append(ret, v)
		}
	}
	return ret
}

func (*slice) UniqueUint32Array(array []uint32) []uint32 {
	checkMap := make(map[uint32]int)
	ret := make([]uint32, 0, len(array))
	for _, v := range array {
		if _, ok := checkMap[v]; !ok {
			ret = append(ret, v)
			checkMap[v] = 1
		}
	}
	return ret
}

func (*slice) UniqueSlice(s1, s2 []uint32) []uint32 {

	res := make([]uint32, 0)

	if len(s1) < 1 {
		return s2
	}

	if len(s2) < 1 {
		return s1
	}

	m := make(map[uint32]bool)

	for _, s := range s1 {

		if _, ok := m[s]; !ok {

			m[s] = true
			res = append(res, s)

		}

	}

	for _, s := range s2 {

		if !m[s] {
			m[s] = true
			res = append(res, s)
		}

	}

	return res

}

func (*slice) DisorderUint32(a []uint32) []uint32 {
	rand.Shuffle(len(a), func(i, j int) {
		a[i], a[j] = a[j], a[i]
	})
	return a
	//if len(a) < 2 {
	//	return a
	//}

	//randVals := make(map[int]uint32)
	//for i, v := range a {
	//	randVals[i] = v
	//}

	//ret := make([]uint32, 0, len(a))
	//for _, v := range randVals {
	//	ret = append(ret, v)
	//}

	//return ret
}

func (*slice) DisorderInt32(a []int32) []int32 {
	rand.Shuffle(len(a), func(i, j int) {
		a[i], a[j] = a[j], a[i]
	})
	return a
	//if len(a) < 2 {
	//	return a
	//}

	//randVals := make(map[int]int32)
	//for i, v := range a {
	//	randVals[i] = v
	//}

	//ret := make([]int32, 0, len(a))
	//for _, v := range randVals {
	//	ret = append(ret, v)
	//}

	//return ret
}

func (*slice) Uint32Sub(a, b []uint32) []uint32 {
	if len(a) == 0 || len(b) == 0 {
		return a
	}

	bMap := make(map[uint32]struct{}, len(b))
	for _, v := range b {
		bMap[v] = struct{}{}
	}

	ret := make([]uint32, 0, len(a))
	for _, v := range a {
		if _, ok := bMap[v]; !ok {
			ret = append(ret, v)
		}
	}

	return ret
}

func (*slice) SplitUint32(a []uint32, num int) [][]uint32 {
	if len(a) == 0 || num <= 0 {
		return nil
	}

	ret := make([][]uint32, 0, len(a)/num+1)

	for i := 0; i < len(a)/num+1; i++ {
		start := num * i
		if start >= len(a) {
			break
		}
		end := start + num
		if end > len(a) {
			end = len(a)
		}
		ret = append(ret, a[start:end])
	}

	return ret
}

// 计算两个slice的交集
func (*slice) Intersect(a, b []uint32) []uint32 {
	if len(a) == 0 || len(b) == 0 {
		return nil
	}
	aSet := gset.New()
	aSet.Add(gconv.Interfaces(a)...)

	bSet := gset.New()
	bSet.Add(gconv.Interfaces(b)...)
	vl := aSet.Intersect(bSet).Slice()
	return gconv.Uint32s(vl)
}

// 通过排序拿到key
func (*slice) SortAndGetKey(arr []uint32) string {
	data := gconv.Ints(arr)
	sort.Ints(data)
	key := "key"
	length := len(data)
	for i := 0; i < length; i++ {
		key = fmt.Sprintf("%s,%d", key, data[i])
	}
	return key
}
