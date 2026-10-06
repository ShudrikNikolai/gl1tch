package main

import "sync"

//type MyMap struct {
//	mu sync.Mutex
//	m  map[string]string
//}
//
//func NewMyMap() *MyMap {
//	return &MyMap{m: make(map[string]string)}
//}
//
//func (mm *MyMap) GerOrCreate(key, value string) string {
//	mm.mu.Lock()
//	defer mm.mu.Unlock()
//	if val, ok := mm.m[key]; ok {
//		return val
//	}
//	mm.m[key] = value
//	return value
//}

// //////////////
var cache sync.Map

func GetOrCompute(key string) string {
	if value, ok := cache.Load(key); ok {
		return value.(string)
	}
	newValue := "1"
	cache.Store(key, newValue)

	return newValue
}

func main() {
	//myNewMap := NewMyMap()
	//myNewMap.GerOrCreate("a", "b")

	/////---------
	//m := map[string][]string{
	//	"tr1": {"tr", "tr1", "tr2", "tr3"},
	//	"tr2": {"tr", "tr2", "tr3", "tr4"},
	//}
	//
	//newVal := []string{"tr1", "tr2", "tr2", "tr2", "tr3", "tr5"}
	//key := "tr1"
	//upd(key, newVal, m)
	//fmt.Println(m[key])

}

func upd(key string, val []string, m map[string][]string) {
	newSlice := append(m[key], val...)
	result := make([]string, 0, len(newSlice))
	mm := make(map[string]struct{}, len(newSlice))
	for _, v := range newSlice {
		if _, ok := mm[v]; !ok {
			mm[v] = struct{}{}
			result = append(result, v)
		}
	}
	m[key] = result
}
