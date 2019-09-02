package xgin

import (
	"encoding/json"
	"reflect"
	"sort"
	"strconv"
)

func getSortSlice(obj interface{}, str string) interface{} {
	data, _ := json.Marshal(obj)
	ms := []map[string]interface{}{}
	json.Unmarshal(data, &ms)
	sort.Slice(ms, func(i, j int) bool {
		if str == "" {
			return false
		}
		field := ""
		asc := true
		if str[:1] == "-" {
			field = str[1:]
			asc = false
		} else {
			field = str
		}
		v1, ok := ms[i][field].(float64)
		if !ok {
			v1 = 0
		}
		v2, ok := ms[j][field].(float64)
		if !ok {
			v2 = 0
		}
		if asc {
			return v1 < v2
		} else {
			return v1 > v2
		}
	})
	return ms
}

func getDecimal2(obj interface{}) interface{} {
	data, _ := json.Marshal(obj)
	ms := []map[string]interface{}{}
	json.Unmarshal(data, &ms)
	for i, m := range ms {
		for k, v := range m {
			if vf, ok := v.(float64); ok && vf != 0 {
				ms[i][k], _ = strconv.ParseFloat(strconv.FormatFloat(vf, 'f', 2, 64), 64)
			}
		}
	}
	return ms
}

// 数组分页
func sliceData(slice interface{}, page, item int) (interface{}, int) {
	fv := reflect.ValueOf(slice)
	if fv.Kind() != reflect.Slice {
		panic("切片分页元数据必须为切片")
	}
	if page == 0 { //第0页返回所有
		return slice, fv.Len()
	}
	data := []interface{}{}
	for i := skip(page, item); i < fv.Len() && i < skip(page, item)+limit(page, item); i++ {
		data = append(data, fv.Index(i).Interface())
	}
	return data, fv.Len()
}
