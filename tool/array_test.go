package tool

import "testing"

func TestSort(t *testing.T) {
	data := make([]map[string]interface{}, 0)
	data = append(data, map[string]interface{}{
		"key": 1,
		//"value":   "303",
		"volume":  67,
		"edition": 2,
	})
	data = append(data, map[string]interface{}{
		"key": 2,
		//"value":   "303",
		"volume":  86,
		"edition": 1,
	})
	data = append(data, map[string]interface{}{
		"key": 3,
		//"value":   "303",
		"volume":  85,
		"edition": 6,
	})
	data = append(data, map[string]interface{}{
		"key": 4,
		//"value":   "303",
		"volume":  98,
		"edition": 2,
	})
	data = append(data, map[string]interface{}{
		"key": 5,
		//"value":   "303",
		"volume":  86,
		"edition": 6,
	})
	data = append(data, map[string]interface{}{
		"key": 6,
		//"value":   "303",
		"volume":  67,
		"edition": 7,
	})
	res := Sort(ArraySort{
		Data: data,
		Match: func(itemI, itemJ map[string]interface{}) bool {
			println(itemI["volume"], itemJ["volume"])
			if itemI["volume"].(int) < itemJ["volume"].(int) {
				return true
			}
			if itemI["volume"].(int) == itemJ["volume"].(int) && itemI["edition"].(int) < itemJ["edition"].(int) {
				return true
			}
			return false
		},
		OrderDesc: true,
	})
	t.Log(res)
}
